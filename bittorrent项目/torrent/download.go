package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"time"
)

type TorrentTask struct {
	PeerId   [20]byte     //我们自己的客户端ID
	PeerList []PeerInfo   //从tracker获取的节点
	InfoSHA  [SHALEN]byte //我们想要获取的文件的唯一标识
	FileName string
	FileLen  int
	PieceLen int
	PieceSHA [][SHALEN]byte //将PieceSHA每20个一组来存放
}

type pieceTask struct { //将大任务拆解成小任务，也就是我们的每个piece
	index  int          //第几个piece
	sha    [SHALEN]byte //对应PieceSHA [][SHALEN]byte的哪一个
	length int          //默认是PieceLen，但是如果是最后一片，就可能短一点
}

type taskState struct { //下载的中间状态
	index int       //第几片
	conn  *PeerConn //跟peer建立的conn是哪一个

	//就是说假如一片有1w bytes，我发发送请求了3000，已经下载了2000
	requested  int //发送了多少请求
	downloaded int //已经下载了多少

	backlog int //并发度
	data    []byte
}

type pieceResult struct { //下载完之后的结果
	index int
	data  []byte
}

const BLOCKSIZE = 16384
const MAXBACKLOG = 5

func (state *taskState) handleMsg() error {
	msg, err := state.conn.ReadMsg()
	if err != nil {
		return err
	}
	// handle keep-alive空消息，保持活跃
	if msg == nil {
		return nil
	}
	switch msg.Id {
	case MsgChoke: //一开始就是默认true，对面不发送。我们握手发送请求之后，对方就会发送请求回来，将他设置为false
		state.conn.Choked = true
	case MsgUnchoke:
		state.conn.Choked = false
	case MsgHave: //告诉对方我已经有了之前没有的piece，需要更新bitmap
		index, err := GetHaveIndex(msg)
		if err != nil {
			return err
		}
		state.conn.Field.SetPiece(index)
	case MsgPiece:
		//返回了数据，拷贝数据到state.data
		n, err := CopyPieceData(state.index, state.data, msg)
		if err != nil {
			return err
		}
		state.downloaded += n //已经下载完了，更新记录
		state.backlog--       //已经下载完了，协程-1
	}
	return nil
}

func downloadPiece(conn *PeerConn, task *pieceTask) (*pieceResult, error) {
	state := &taskState{ //新建一个下载过程的结构体
		index: task.index,
		conn:  conn,
		data:  make([]byte, task.length),
	}
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetDeadline(time.Time{})

	for state.downloaded < task.length {
		if !conn.Choked { //对方节点一定是unchoked，是上传的，如果对方是choke状态，就一直等
			for state.backlog < MAXBACKLOG && state.requested < task.length {
				//并发度不可大于最大并发度&&发送的请求数少于所需的task数
				length := BLOCKSIZE //我们一次发过去的请求长度
				if task.length-state.requested < length {
					//假如是最后一片，那么length就会不足BLOCKSIZE，需要处理边界
					length = task.length - state.requested
				}
				//下载第几个piece，从第几个开始下载，下载的长度
				msg := NewRequestMsg(state.index, state.requested, length)
				_, err := state.conn.WriteMsg(msg)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += length
			}
		}
		err := state.handleMsg()
		if err != nil {
			return nil, err
		}
	}
	return &pieceResult{state.index, state.data}, nil
}

func checkPiece(task *pieceTask, res *pieceResult) bool {
	sha := sha1.Sum(res.data)
	if !bytes.Equal(task.sha[:], sha[:]) {
		fmt.Printf("check integrity failed, index :%v\n", res.index)
		return false
	}
	return true
}

func (t *TorrentTask) peerRoutine(peer PeerInfo, taskQueue chan *pieceTask, resultQueue chan *pieceResult) {
	// set up conn with peer建立peer连接
	conn, err := NewConn(peer, t.InfoSHA, t.PeerId)
	if err != nil {
		fmt.Println("fail to connect peer : " + peer.Ip.String())
		return
	}
	defer conn.Close()

	fmt.Println("complete handshake with peer : " + peer.Ip.String())
	conn.WriteMsg(&PeerMsg{MsgInterested, nil}) //我们想下载的内容
	// get piece task & download
	for task := range taskQueue {
		if !conn.Field.HasPiece(task.index) { //看看我们连接的peer是否有我们的piece
			taskQueue <- task //只要有任何原因，task下载失败，就让这个task回到channel里面，让其他peer下载
			continue
		}
		fmt.Printf("get task, index: %v, peer : %v\n", task.index, peer.Ip.String())
		res, err := downloadPiece(conn, task) //如果包含了就开始下载
		if err != nil {
			taskQueue <- task //只要下载失败，就重新放回channel
			fmt.Println("fail to download piece" + err.Error())
			return
		}
		if !checkPiece(task, res) {
			taskQueue <- task //校验，校验失败也要放回channel
			continue
		}
		resultQueue <- res
	}
}

func (t *TorrentTask) getPieceBounds(index int) (bengin, end int) {
	bengin = index * t.PieceLen
	end = bengin + t.PieceLen
	if end > t.FileLen {
		end = t.FileLen
	}
	return
}

func Download(task *TorrentTask) error {
	fmt.Println("start downloading " + task.FileName)
	// split pieceTasks and init task&result channel//起了两个channel
	taskQueue := make(chan *pieceTask, len(task.PieceSHA)) //channel的容量跟piece的个数一样
	resultQueue := make(chan *pieceResult)                 //容量为1：下载完一个就校验一个，校验一个就写进去一个
	//将大task拆分成小piece
	for index, sha := range task.PieceSHA {
		begin, end := task.getPieceBounds(index) //如果是最后一个，那么它的end可能小于length
		taskQueue <- &pieceTask{index, sha, (end - begin)}
	}
	// init goroutines for each peer，为一个peer新建一个协程
	for _, peer := range task.PeerList {
		go task.peerRoutine(peer, taskQueue, resultQueue) //这里已经做了校验，才写进resultQueue
	}
	// collect piece result
	buf := make([]byte, task.FileLen) //buf用于存数据
	count := 0
	for count < len(task.PieceSHA) { //如果获取的个数小于piece的总数，就继续获取
		res := <-resultQueue
		begin, end := task.getPieceBounds(res.index)
		copy(buf[begin:end], res.data)
		count++
		// print progress
		percent := float64(count) / float64(len(task.PieceSHA)) * 100
		fmt.Printf("downloading, progress : (%0.2f%%)\n", percent)
	}
	close(taskQueue)
	close(resultQueue)
	// create file & copy data
	file, err := os.Create(task.FileName) //buf写进文件里面
	if err != nil {
		fmt.Println("fail to create file: " + task.FileName)
		return err
	}
	_, err = file.Write(buf)
	if err != nil {
		fmt.Println("fail to write data")
		return err
	}
	return nil
}
