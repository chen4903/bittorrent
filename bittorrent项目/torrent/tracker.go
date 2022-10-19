package torrent

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/archeryue/go-torrent/bencode"
)

const (
	PeerPort int = 6666
	IpLen    int = 4
	PortLen  int = 2
	PeerLen  int = IpLen + PortLen
)

const IDLEN int = 20

type PeerInfo struct {
	Ip   net.IP
	Port uint16
}

type TrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func buildUrl(tf *TorrentFile, peerId [IDLEN]byte) (string, error) {
	//这里是不是少了一部分？？？
	base, err := url.Parse(tf.Announce)
	if err != nil {
		fmt.Println("Announce Error: " + tf.Announce)
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(tf.InfoSHA[:])}, //文件标识
		"peer_id":    []string{string(peerId[:])},     //下载器标识
		"port":       []string{strconv.Itoa(PeerPort)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tf.FileLen)}, //文件的剩余大小
	}

	base.RawQuery = params.Encode() //进行URL编码
	return base.String(), nil
}

func buildPeerInfo(peers []byte) []PeerInfo {
	//该函数：获得peers的IP和port，紧凑排列
	num := len(peers) / PeerLen
	if len(peers)%PeerLen != 0 {
		fmt.Println("Received malformed peers")
		return nil
	}
	infos := make([]PeerInfo, num)
	for i := 0; i < num; i++ {
		offset := i * PeerLen
		infos[i].Ip = net.IP(peers[offset : offset+IpLen])
		infos[i].Port = binary.BigEndian.Uint16(peers[offset+IpLen : offset+PeerLen])
	}
	return infos
}

func FindPeers(tf *TorrentFile, peerId [IDLEN]byte) []PeerInfo {
	//请求
	url, err := buildUrl(tf, peerId)
	if err != nil {
		fmt.Println("Build Tracker Url Error: " + err.Error())
		return nil
	}

	//HTTP的get
	cli := &http.Client{Timeout: 15 * time.Second}
	resp, err := cli.Get(url) //resp是bencode编码的结构体
	if err != nil {
		fmt.Println("Fail to Connect to Tracker: " + err.Error())
		return nil
	}
	defer resp.Body.Close()

	trackResp := new(TrackerResp)
	err = bencode.Unmarshal(resp.Body, trackResp)
	if err != nil {
		fmt.Println("Tracker Response Error" + err.Error())
		return nil
	}

	return buildPeerInfo([]byte(trackResp.Peers))
}
