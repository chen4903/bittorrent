# GO手写BT下载器

项目：https://github.com/archeryue/go-postern

项目：https://blog.jse.li/posts/torrent/

## 简介

### Torrent File格式

- announce：string =>URL
- announce-list：[ ]string =>备用的tracker列表
- info（dict）：文件具体信息（单文件）
  - name：string。比如：《无间道》
  - length：int => 文件总长度。比如：3.5G
  - piceces： [20]byte  每个文件片的SHA-1值。用于检验文件片
  - picece length：int => 每个文件片的长度。比如：100MB
- 多文件：
  - name、length、piceces、picece length
  - files：它是一个list
    - path：文件路径。第一级
    - length：文件长度。第一级
    - ...
    - path。第n级
    - length。第n级

### Bencode编码

**Bencode**使用ASCII字符进行编码。比Json更适合，因为它开头的第一个标识符就确定了传输内容的类型

- **整型数int**：一个整型数int会以十进制数编码并括在i和e之间，不允许前导零（但0依然为整数0），负数如十进制表示一样使用前导负号，不允许负零。如整型数“42”编码为“`i42e`”，数字“0”编码为“`i0e`”，“-42”编码为“`i-42e`”。
- **字符串string**：一个以字节为单位表示的字符串string（字符串的字为一个字节，不一定是一个字符）会以`（长度）:（内容）`编码，长度的值和数字编码方法一样，只是不允许负数；内容就是字符串的内容，如字符串“spam”就会编码为“`4:spam`”，本规则不能处理ASCII以外的字符串，为了解决这个问题，一些BitTorrent程序会以非标准的方式将ASCII以外的字符以UTF-8编码转化后再编码。
- **线性表list**：会以l和e包住来编码，其中的内容为Bencode四种编码格式所组成的编码字符串，如包含和字符串“spam”数字“42”的线性表会被编码为“`l4:spami42ee`”，注意分隔符要对应配对。（注意：是第一个符号是小写的L，而不是14）
- **字典表dict**：会以d和e包住来编码，字典元素的键和值必须紧跟在一起，而且所有键为字符串类型并按字典顺序排好。如键为“bar”值为字符串“spam”和键为“foo”值为整数“42”的字典表会被编码为“`d3:bar4:spam3:fooi42ee`”。

### 学习计划

1. Bencode库：用于做Bencode的序列化与反序列化
2. torrent解析库：针对任何一个torrent文件都可以解析出来内容。这一步解析出tracker信息和种子的信息
3. tracker模块：获取peers信息
4. download模块：让peers进行交互，piceces下载和校验
5. assembeer模块：拼装文件片成file

## 项目编写

### 第一部分：Bencode库

为了方便，我们将Bencode支持的四种数据结构称为：`BObject`

在Go项目中，从io.Reader读取[ ]Bytes，然后解析成BObject，最后io.writer解码写回

### 第二部分：torrent文件解析

XXX.torrent====>struct，这一步是Unmarshal

struct、slice=====>XXX.torrent，这一步是marshal

### 第三部分

### 第四部分：Peer通信机制

TCP协议，握手消息一共68byte（1 + 19 +8 + 20 + 20）

- 握手：交换信息（bit Torrent，peerID，InfoSHA）
- 获取peer数据情况
- 指定piece下载

![](imags\握手信息.png)









