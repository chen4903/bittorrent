package bencode

import (
	"bufio"
	"io"
)

func Parse(r io.Reader) (*BObject, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	//recursive descent parsing
	b, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	var ret BObject
	switch { //看一眼*BObject的第一位是什么，不取出，然后做相应的处理
	case b[0] >= '0' && b[0] <= '9':
		// parse string
		val, err := DecodeString(br)
		if err != nil {
			return nil, err
		}
		ret.type_ = BSTR
		ret.val_ = val
	case b[0] == 'i':
		// parse int
		val, err := DecodeInt(br)
		if err != nil {
			return nil, err
		}
		ret.type_ = BINT
		ret.val_ = val
	case b[0] == 'l':
		// parse list
		br.ReadByte()
		var list []*BObject
		for { //list可能有若干个BObject
			if p, _ := br.Peek(1); p[0] == 'e' {
				//不断peek，如果是e，就说明读到了最后一位了
				br.ReadByte()
				break
			}
			elem, err := Parse(br) //递归调用，因为每个BObject可能是list或dict
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		ret.type_ = BLIST
		ret.val_ = list
	case b[0] == 'd':
		// parse map
		br.ReadByte()
		dict := make(map[string]*BObject)
		for {
			if p, _ := br.Peek(1); p[0] == 'e' {
				br.ReadByte()
				break
			}
			key, err := DecodeString(br) //先把dict的key解析出来
			if err != nil {
				return nil, err
			}
			val, err := Parse(br) //递归调用，因为每个BObject可能是list或dict
			if err != nil {
				return nil, err
			}
			dict[key] = val
		}
		ret.type_ = BDICT
		ret.val_ = dict
	default:
		return nil, ErrIvd
	}
	return &ret, nil
}
