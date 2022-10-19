package bencode

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrNum = errors.New("expect num")
	ErrCol = errors.New("expect colon")
	ErrEpI = errors.New("expect char i")
	ErrEpE = errors.New("expect char e")
	ErrTyp = errors.New("wrong type")
	ErrIvd = errors.New("invalid bencode")
)

type BType uint8 //因为只有四种可能性，所以8位足够

const (
	BSTR  BType = 0x01
	BINT  BType = 0x02
	BLIST BType = 0x03
	BDICT BType = 0x04
)

type BValue interface{} //空接口：BValue因为有多种形式，所以用接口类型（因为没有泛型支持）

type BObject struct { //一个BObject里面有对应的key-value关系
	type_ BType
	val_  BValue
}

func (o *BObject) Str() (string, error) { //如果传入的是string就返回string
	if o.type_ != BSTR {
		return "", ErrTyp
	}
	return o.val_.(string), nil
}

func (o *BObject) Int() (int, error) { //如果传入的是int就返回int
	if o.type_ != BINT {
		return 0, ErrTyp
	}
	return o.val_.(int), nil
}

func (o *BObject) List() ([]*BObject, error) {
	if o.type_ != BLIST {
		return nil, ErrTyp
	}
	//返回的是指针的slice，因为BObject可能很大
	return o.val_.([]*BObject), nil
}

func (o *BObject) Dict() (map[string]*BObject, error) {
	if o.type_ != BDICT {
		return nil, ErrTyp
	}
	return o.val_.(map[string]*BObject), nil
}

func (o *BObject) Bencode(w io.Writer) int {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	wLen := 0
	switch o.type_ {
	case BSTR:
		str, _ := o.Str()
		wLen += EncodeString(bw, str)
	case BINT:
		val, _ := o.Int()
		wLen += EncodeInt(bw, val)
	case BLIST:
		bw.WriteByte('l')
		list, _ := o.List() //list是一个切片[]
		for _, elem := range list {
			wLen += elem.Bencode(bw) //递归调用：因为切片里面的值可能还是list或dict
		}
		bw.WriteByte('e')
		wLen += 2
	case BDICT:
		bw.WriteByte('d')
		dict, _ := o.Dict() //dict是一个map[]
		for k, v := range dict {
			wLen += EncodeString(bw, k) //将key写进缓存
			wLen += v.Bencode(bw)       //value递归遍历写入缓存，因为value可能是list或dict
		}
		bw.WriteByte('e')
		wLen += 2
	}
	bw.Flush()
	return wLen
}

func checkNum(data byte) bool {
	return data >= '0' && data <= '9'
}

func readDecimal(r *bufio.Reader) (val int, len int) {
	//这个方法是将'1''9''9'转化成199
	sign := 1
	b, _ := r.ReadByte()
	len++
	if b == '-' {
		sign = -1
		b, _ = r.ReadByte()
		len++
	}
	for {
		if !checkNum(b) { //如果读取到的下一个不是数字了，就退出读取
			r.UnreadByte()
			len--
			return sign * val, len
		}
		val = val*10 + int(b-'0')
		b, _ = r.ReadByte()
		len++
	}
}

func writeDecimal(w *bufio.Writer, val int) (len int) {
	//以数字199为例子，将199转化成对应的ASC｜｜码值
	if val == 0 {
		w.WriteByte('0')
		len++
		return
	}
	if val < 0 { //如果这个值是负数，就加上负号
		w.WriteByte('-')
		len++
		val *= -1 //同时将这个负值转化成正的
	}

	dividend := 1
	//199为例子：取出1：199　/　100
	//　　　　　　取出9：199　/　10
	//　　　　　　取出9：199　/　1
	//那么如何找到这个100呢？将1不断乘10，当这个数大于199的时候，除以10，就是目标值了
	//这里就是1000　/　10　＝　100
	for {
		if dividend > val {
			dividend /= 10
			break
		}
		dividend *= 10
	}
	for {
		num := byte(val / dividend)
		//因为不可以直接将这个num存进去，加上0的ASC||偏移量，就可以找到数字对应的ASC值
		w.WriteByte('0' + num)
		len++
		if dividend == 1 {
			return
		}
		val %= dividend
		dividend /= 10
	}
}

func EncodeString(w io.Writer, val string) int {
	//比如："abc"  ===>  3:abc
	strLen := len(val)               //得到字符串val的长度
	bw := bufio.NewWriter(w)         //将w写入缓存区
	wLen := writeDecimal(bw, strLen) //将这个数字（3）以十进制数字对应的ASCII码值写进缓存
	bw.WriteByte(':')                //给数字后面加个“:”
	wLen++                           //缓存区bw内容，加上冒号的长度
	bw.WriteString(val)              //将val（abc）写进缓存
	wLen += strLen                   //wLen的值：解析后字符串的长度，如：3:abc

	err := bw.Flush() //3:abc已经写进了缓存区，FLush一下
	if err != nil {
		return 0
	}
	return wLen
}

func DecodeString(r io.Reader) (val string, err error) {
	//比如：3:abc  ===>  "abc"
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	//将io.Reader中3:abc的数字部分取出来
	//num是abc的长度3，len是1
	num, len := readDecimal(br)
	if len == 0 {
		return val, ErrNum
	}
	b, err := br.ReadByte() //将io.Reader中3:abc的冒号取出来
	if b != ':' {
		return val, ErrCol
	}
	buf := make([]byte, num) //创建一个长度为3的切片，用于将abc放进去
	//在io.writer中读取剩余的内容，读取的数量为num，放到buf中
	_, err = io.ReadAtLeast(br, buf, num)
	val = string(buf)
	return
}

func EncodeInt(w io.Writer, val int) int {
	bw := bufio.NewWriter(w)
	wLen := 0
	bw.WriteByte('i')
	wLen++
	nLen := writeDecimal(bw, val) //将这个数字转换成ASC||值写进缓存
	wLen += nLen
	bw.WriteByte('e')
	wLen++

	err := bw.Flush()
	if err != nil {
		return 0
	}
	return wLen
}

func DecodeInt(r io.Reader) (val int, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	b, err := br.ReadByte()
	if b != 'i' {
		return val, ErrEpI
	}
	val, _ = readDecimal(br)
	b, err = br.ReadByte()
	if b != 'e' {
		return val, ErrEpE
	}
	return
}
