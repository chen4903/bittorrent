# Kademlia

Kademlia是一种通过分散式杂凑表实现的协议算法，它是由Petar Maymounkov与David Mazières为非集中式P2P计算机网络而设计的。Kademlia规定了网络的结构，也规定了通过节点查询进行信息交换的方式。

参与通讯的所有节点形成一张虚拟网（或者叫做覆盖网）。这些节点通过一组数字（或称为节点ID）来进行身份标识。节点ID不仅可以用来做身份标识，还可以用来进行值定位（值通常是文件的散列或者关键词）。其实，节点ID与文件散列直接对应，它所表示的那个节点存储着哪儿能够获取文件和资源的相关信息。

当我们在网络中搜索某些值（即通常搜索存储文件散列或关键词的节点）的时候，Kademlia算法需要知道与这些值相关的键，然后分步在网络中开始搜索。每一步都会找到一些节点，这些节点的ID与键更为接近，如果有节点直接返回搜索的值或者再也无法找到与键更为接近的节点ID的时候搜索便会停止。这种搜索值的方法是非常高效的：与其他的分散式杂凑表的实现类似，在一个包含n个节点的系统的值的搜索中，Kademlia仅访问O(log(n))个节点。

非集中式网络结构还有更大的优势，那就是它能够显著增强抵御拒绝服务攻击的能力。即使网络中的一整批节点遭受泛洪攻击，也不会对网络的可用性造成很大的影响，通过绕过这些漏洞（被攻击的节点）来重新编织一张网络，网络的可用性就可以得到恢复。

## 系统细节

第一代P2P文件分享网络，像Napster，依赖于中央数据库来协调网络中的查询，第二代P2P网络，像Gnutella，使用泛滥式查询（query flooding）来查询文件，它会搜索网络中的所有节点，第三代p2p网络使用分散式杂凑表来查询网络中的文件，分散式杂凑表在整个网络中储存资源的位置，这些协议追求的主要目标就是快速定位期望的节点。

Kademlia基于两个节点之间的**距离计算**，该距离是两个网络节点ID号的异或（ XOR distance ），计算的结果最终作为整型数值返回。关键字和节点ID有同样的格式和长度，因此，可以使用同样的方法计算关键字和节点ID之间的距离。节点ID一般是一个大的随机数，选择该数的时候所追求的一个目标就是它的唯一性（希望在整个网络中该节点ID是唯一的）。异或距离跟实际上的地理位置没有任何关系，只与ID相关。因此很可能来自中国和阿根廷的节点由于选择了相似的随机ID而成为邻居。选择异或是因为通过它计算的距离享有几何距离公式的一些特征，尤其体现在以下几点：节点和它本身之间的异或距离是0；异或距离是对称的：即从A到B的异或距离与从B到A的异或距离是等同的；**异或距离符合三角不等式**：三个顶点A B C，AC异或距离小于或等于AB异或距离和BC异或距离之和。由于以上的这些属性，在实际的节点距离的度量过程中计算量将大大降低。Kademlia搜索的每一次迭代将距目标至少更近1 bit。一个基本的具有2的n次方个节点的Kademlia网络在最坏的情况下只需花n步就可找到被搜索的节点或值。

## 路由表

为了说明简单，本部分基于单个bit构建路由表，如需关于实际路由表的更多信息，请看“查询加速”部分。

**Kademlia**路由表由多个列表组成，每个列表对应节点ID的一位（例如：假如节点ID共有128位，则节点的路由表将包含128个列表），包含多个条目，条目中包含定位其他节点所必要的一些数据。列表条目中的这些数据通常是由其他节点的IP地址，端口和节点ID组成。每个列表对应于与节点相距特定范围距离的一些节点，节点的第n个列表中所找到的节点的第n位与该节点的第n位肯定不同，而前n-1位相同，这就意味着很容易使用网络中远离该节点的一半节点来填充第一个列表（第一位不同的节点最多有一半），而用网络中四分之一的节点来填充第二个列表（比第一个列表中的那些节点离该节点更近一位），依次类推。如果ID有128个二进制位，则网络中的每个节点按照不同的异或距离把其他所有的节点分成了128类，ID的每一位对应于其中的一类。随着网络中的节点被某节点发现，它们被逐步加入到该节点的相应的列表中，这个过程中包括向节点列表中存信息和从节点列表中取信息的操作，甚至还包括当时协助其他节点寻找相应键对应值的操作。这个过程中发现的所有节点都将被加入到节点的列表之中，因此节点对整个网络的感知是动态的，这使得网络一直保持着频繁地更新，增强了抵御错误和攻击的能力。

在**Kademlia**相关的论文中，列表也称为K桶，其中K是一个系统变量，如20，每一个K桶是一个最多包含K个条目的列表，也就是说，网络中所有节点的一个列表（对应于某一位，与该节点相距一个特定的距离）最多包含20个节点。随着对应的bit位变低（即对应的异或距离越来越短），K桶包含的可能节点数迅速下降（这是由于K桶对应的异或距离越近，节点数越少），因此，对应于更低bit位的K桶显然包含网络中所有相关部分的节点。由于网络中节点的实际数量远远小于可能ID号的数量，所以对应那些短距离的某些K桶可能一直是空的（如果异或距离只有1，可能的数量就最大只能为1，这个异或距离为1的节点如果没有发现，则对应于异或距离为1的K桶则是空的）。

让我们看下面的那个简单网络，该网络最大可有2^3，即8个关键字和节点，目前共有7个节点加入，每个节点用一个小圈表示（在树的底部）。我们考虑那个用黑圈标注的节点6，它共有3个K桶，节点0，1和2（二进制表示为000，001和010）是第一个K桶的候选节点，节点3目前（二进制表示为011）还没有加入网络，节点4和节点5（二进制表示分别为100和101）是第二个K桶的候选节点，只有节点7（二进制表示为111）是第3个K桶的候选节点。图中，3个K桶都用灰色圈表示，假如K桶的大小（即K值）是2，那么第一个K桶只能包含3个节点中的2个。众所周知，那些长时间在线连接的节点未来长时间在线的可能性更大，基于这种静态统计分布的规律，**Kademlia**选择把那些长时间在线的节点存入K桶，这一方法增长了未来某一时刻有效节点的数量，同时也提供了更为稳定的网络。当某个K桶已满，而又发现了相应于该桶的新节点的时候，那么，就首先检查K桶中最早访问的节点，假如该节点仍然存活，那么新节点就被安排到一个附属列表中（作为一个替代缓存）.只有当K桶中的某个节点停止响应的时候，替代cache才被使用。换句话说，新发现的节点只有在老的节点消失后才被使用。

![110节点的网络分区](https://upload.wikimedia.org/wikipedia/commons/thumb/6/63/Dht_example_SVG.svg/630px-Dht_example_SVG.svg.png)

## 协议消息

**Kademlia**协议共有四种消息。

- PING消息：用来测试节点是否仍然在线。
- STORE消息：在某个节点中存储一个键值对
- FIND_NODE消息：消息请求的接收者将返回自己桶中离请求键值最近的K个节点。
- FIND_VALUE消息：与FIND_NODE一样，不过当请求的接收者存有请求者所请求的键的时候，它将返回相应键的值。每一个RPC消息中都包含一个发起者加入的随机值，这一点确保响应消息在收到的时候能够与前面发送的请求消息匹配。

## 定位节点

节点查询可以异步进行，也可以同时进行，同时查询的数量由α表示，一般是3。在节点查询的时候，它先得到它K桶中离所查询的键值最近的K个节点，然后向这K个节点发起FIND_NODE消息请求，消息接收者收到这些请求消息后将在他们的K桶中进行查询，如果他们知道离被查键更近的节点，他们就返回这些节点（最多K个）。消息的请求者在收到响应后将使用它所收到的响应结果来更新它的结果列表，这个结果列表总是保持K个响应FIND_NODE消息请求的最优节点（即离被搜索键更近的K个节点）。然后消息发起者将向这K个最优节点发起查询，不断地迭代执行上述查询过程。因为每一个节点比其他节点对它周边的节点有更好的感知能力，因此响应结果将是一次一次离被搜索键值越来越近的某节点。如果本次响应结果中的节点没有比前次响应结果中的节点离被搜索键值更近了，这个查询迭代也就终止了。当这个迭代终止的时候，响应结果集中的K个最优节点就是整个网络中离被搜索键值最近的K个节点（从以上过程看，这显然是局部的，而非整个网络）。

节点信息中可以增加一个往返时间，或者叫做RTT的参数，这个参数可以被用来定义一个针对每个被查询节点的超时设置，即当向某个节点发起的查询超时的时候，另一个查询才会发起，当然，针对某个节点的查询在同一时刻从来不超过α个。

## 定位资源

通过把资源信息与键进行映射，资源即可进行定位，杂凑表是典型的用来映射的手段。由于以前的STORE消息，存储节点将会有对应STORE所存储的相关资源的信息。定位资源时，如果一个节点存有相应的资源的值的时候，它就返回该资源，搜索便结束了，除了该点以外，定位资源与定位离键最近的节点的过程相似。

考虑到节点未必都在线的情况，资源的值被存在多个节点上（节点中的K个），并且，为了提供冗余，还有可能在更多的节点上储存值。储存值的节点将定期搜索网络中与储存值所对应的键接近的K个节点并且把值复制到这些节点上，这些节点可作为那些下线的节点的补充。另外，对于那些普遍流行的内容，可能有更多的请求需求，通过让那些访问值的节点把值存储在附件的一些节点上（不在K个最近节点的范围之类）来减少存储值的那些节点的负载，这种新的存储技术就是缓存技术。通过这种技术，依赖于请求的数量，资源的值被存储在离键越来越远的那些节点上，这使得那些流行的搜索可以更快地找到资源的储存者。由于返回值的节点的NODE_ID远离值所对应的关键字（个人理解：就是说提供值的人的ID和热点信息映射出来的ID差太远了，并不会被归类在一起），网络中的“热点”区域存在的可能性也降低了。依据与键的距离，缓存的那些节点在一段时间以后将会删除所存储的缓存值。分散式杂凑表的某些实现（如Kad）即不提供冗余（复制）节点也不提供缓存，这主要是为了能够快速减少系统中的陈旧信息。在这种网络中，提供文件的那些节点将会周期性地更新网络上的信息（通过FIND_NODE消息和STORE消息）。当存有某个文件的所有节点都下线了，关于该文件的相关的值（源和关键字）的更新也就停止了，该文件的相关信息也就从网络上完全消失了。

## 加入网络

想要加入网络的节点首先要经历一个引导过程。在引导过程中，节点需要知道其他已加入该网络的某个节点的IP地址和端口号（可从用户或者存储的列表中获得）。假如正在引导的那个节点还未加入网络，它会计算一个目前为止还未分配给其他节点的随机ID号，直到离开网络，该节点会一直使用该ID号。

以下是新加入的节点如何定位到ｋ桶：

正在加入**Kademlia**网络的节点在它的某个K桶中插入引导节点（负责加入节点的初始化工作），然后向它的唯一邻居（引导节点）发起FIND_NODE操作请求来定位自己，这种“自我定位”将使得**Kademlia**的其他节点（收到请求的节点）能够使用新加入节点的Node Id填充他们的K桶，同时也能够使用那些查询过程的中间节点（位于新加入节点和引导节点的查询路径上的其他节点）来填充新加入节点的K桶。这一自查询过程使得新加入节点自引导节点所在的那个K桶开始，由远及近，逐个得到刷新，这种刷新只需通过位于K桶范围内的一个随机键的定位便可达到。

以下是Ｋ桶的分裂过程：

最初的时候，节点仅有一个K桶（覆盖所有的ID范围），当有新节点需要插入该K桶时，如果K桶已满，K桶就开始分裂，（参见APeer-to-peer Information System 2.4）分裂发生在节点的K桶的覆盖范围（表现为二叉树某部分从左至右的所有值）包含了该节点本身的ID的时候。对于节点内距离节点最近的那个K桶，**Kademlia**可以放松限制（即可以到达K时不发生分裂），因为桶内的所有节点离该节点距离最近，这些节点个数很可能超过K个，而且节点希望知道所有的这些最近的节点。因此，在路由树中，该节点附近很可能出现高度不平衡的二叉子树。假如K是20，新加入网络的节点ID为“xxx000011001”，则前缀为“xxx0011……”的节点可能有21个，甚至更多，新的节点可能包含多个含有21个以上节点的K桶。（位于节点附近的k桶）。这点保证使得该节点能够感知网络中附近区域的所有节点。（参见A Peer-to-peer Information System 2.4）

## 查询加速

**Kademlia**使用异或来定义距离。两个节点ID的异或（或者节点ID和关键字的异或）的结果就是两者之间的距离。对于每一个二进制位来说，如果相同，异或返回0，否则，异或返回1。异或距离满足三角形不等式：任何一边的距离小于（或等于）其它两边距离之和。

异或距离使得**Kademlia**的路由表可以建在单个bit之上，即可使用位组（多个位联合）来构建路由表。位组可以用来表示相应的K桶，它有个专业术语叫做前缀，对一个m位的前缀来说，可对应2^m-1个K桶。（m位的前缀本来可以对应2^m个K桶）另外的那个K桶可以进一步扩展为包含该节点本身ID的路由树。一个b位的前缀可以把查询的最大次数从logn减少到logn/b。这只是查询次数的最大值，因为自己K桶可能比前缀有更多的位与目标键相同，（这会增加在自己K桶中找到节点的机会，假设前缀有m位，很可能查询一个节点就能匹配2m甚至更多的位组），所以其实平均的查询次数要少的多。（参考Improving Lookup Performance over a Widely-DeployedDHT第三部分）

节点可以在他们的路由表中使用混合前缀，就像eMule中的Kad网络。如果以增加查询的复杂性为代价，**Kademlia**网络在路由表的具体实现上甚至可以是有异构的。

## 在文件分享网络中的应用

**Kademlia**可在文件分享网络中使用，通过制作**Kademlia**关键字搜索，我们能够在文件分享网络中找到我们需要的文件以供我们下载。由于没有中央服务器存储文件的索引，这部分工作就被平均地分配到所有的客户端中去：假如一个节点希望分享某个文件，它先根据文件的内容来处理该文件，通过运算，把文件的内容散列成一组数字，该数字在文件分享网络中可被用来标识文件。这组散列数字必须和节点ID有同样的长度，然后，该节点便在网络中搜索ID值与文件的散列值相近的节点，并把它自己的IP地址存储在那些搜索到的节点上，也就是说，它把自己作为文件的源进行了发布。正在进行文件搜索的客户端将使用**Kademlia**协议来寻找网络上ID值与希望寻找的文件的散列值最近的那个节点，然后取得存储在那个节点上的文件源列表。

由于一个键可以对应很多值，即同一个文件可以有多个源，每一个存储源列表的节点可能有不同的文件的源的信息，这样的话，源列表可以从与键值相近的K个节点获得。 文件的散列值通常可以从其他的一些特别的Internet链接的地方获得，或者被包含在从其他某处获得的索引文件中。

文件名的搜索可以使用关键词来实现，文件名可以分割成连续的几个关键词，这些关键词都可以散列并且可以和相应的文件名和文件散列储存在网络中。搜索者可以使用其中的某个关键词，联系ID值与关键词散列最近的那个节点，取得包含该关键词的文件列表。由于在文件列表中的文件都有相关的散列值，通过该散列值就可利用上述通常取文件的方法获得要搜索的文件。

## 大白话例子

这么多文字，其实很难了解到一个大概过程，那么下面我将基于[回形针的关于BT种子的视频](https://www.youtube.com/watch?v=jp0bF9Qu2Jw&t=316s)来对本技术进行一个对应本文章的具体文字阐述。本例子假设本人是`0100`节点，目标资源只有一个节点拥有。

（１）

1. 根据现有的节点情况，将所有的节点根据二进制ID划分成二叉树的各个叶子节点。
2. 然后，根据自己不在哪一半边的子树来划分下一个K桶（比如：我是`0100`，我不在`010？`的左边子树，就将左边子树设置为0号K桶。之后向上继续划分，我是`010？`这边的，那么把`01？？`左边的子树划分成下一个K桶，即1号K桶……）。
3. 按照这样的方式，将整个二叉树划分出了n个K桶（这个n就是ID一共n位）。

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/%E5%9B%9E%E5%BD%A2%E9%92%88_K%E6%A1%B6%E5%88%92%E5%88%86.png?raw=true)

（2）

1. 先建立一个路由表：根据前缀来计算二进制距离
   1. 每一个K桶里面都有一个固定的前缀（即？前面的数字），他有多个可能的值，这个值减去我的值，就可以得到一个距离范围
   2. 以此类推，就可以得到各自的距离范围
2. 图中的节点只是取了两个，而实际情况可能有很多个，视频是为了简化过程。
3. 如果按照图中的节点，就是说只有除了本身之外，还有7个节点在线

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/%E5%9B%9E%E5%BD%A2%E9%92%88_%E8%B7%AF%E7%94%B1%E8%A1%A8.png?raw=true)

（3）

1. 我的节点是0100，目标节点是1111，异或出来的二进制距离是1011，也就是11，在上图中查询，发现是在3号K桶。这说明我们要找的节点肯定是在以0100划分的3号K桶之中
2. 3号K桶中，取一个在线的节点比如`1110`，根据`1110`来再次划分K桶

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/%E5%9B%9E%E5%BD%A2%E9%92%88_%E8%8A%82%E7%82%B9K%E6%A1%B6%E5%86%8D%E5%88%92%E5%88%86.png?raw=true)

（4）

1. 本处以上图`1110`来划分K桶
2. 视频中，是直接取了0号K桶来进行异或（1110异或1111为0001，即1），发现距离结果是1，说明0号K桶中的`1111`就是目标节点
3. 但是，实际情况并不总是这么巧，很有可能目标节点是在2号K桶中的`1011`呢？（我们假如`1011`就是目标节点）
4. 那么我们就要分情况了。现在假设0号K桶的`1111`节点在线，1号K桶的`1101`和`1100`节点在线，2号K桶中仅有一个节点`1010`在线
   1. 将`1110`以上四个节点ID进行异或，得出二进制距离，发现最短的是`1010`节点（距离不等于1）。然后对照以1110划分出来的路由表，我们可以定位到2号K桶
   2. 但是此时2号K桶除了`1010`节点，没有其他节点在线了，说明资源拥有者并不在线。
   3. 此时我们就可以将资源的拥有者定位到了：以`1110`来划分的2号K桶之中
5. 按照步骤4的方法，对不同的在线节点循环迭代，就可以一步步缩小范围。如果资源拥有者在线，就一定可以定位到他，如果资源拥有者不在线，那就会尽可能地缩小范围定位到资源拥有者的位置，等待它的上线

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/%E5%9B%9E%E5%BD%A2%E9%92%88_%E8%B7%AF%E7%94%B1%E8%A1%A8_.png?raw=true)

## 进一步理解

这里我就不翻译了，我是根据视频直接输入英文上去的

### kademlia算法概述

#### topics

1. Elements of Kademlia node
2. Network initialization
3. Kademlia API
4. Storing data in Kademlia nodes

#### Elements of Kademlia node

- GUID
- Address of node
- Routing Table
- API to call other nodes

a GUID is unique on the identifier for each node that is in certain bit space.

if our bit space is 4 we get maximum 2 to power of 4 guids which means we have node guid from 0 to 15.

Address o node: it could be basically an ip address and a port or it could be an http address or whatever else.

at last each node has an API to accept imcoming requests.

#### Routing Table

- Contains List of Buckets
- Bucket List size is same as network GUID space
- List<Bucket<Node>>

the number of pockets we have is same as our network guide space.

nodes that can go in the ends list must have a differing end bit from the node's id

the first n-1 bits of the candidate id must match those of the node's id

this means that it is very easy to populate the first list as half of the nodes in the network are far away from the candidates. the next list can use only quarter of the nodes

##### step1

the 0 is guid of the node, then our second node which is node 1 wants to join the network

since node 1 could get response from node 0

it adds node 0 to its own loading table note that i'm using the word close

node 0 is a bootstrap node since it was used by node 1 to build its initial routing table 

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/%E8%8A%82%E7%82%B91%E6%83%B3%E5%8A%A0%E5%85%A5.png?raw=true)

##### step2

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/%E8%8A%82%E7%82%B92%E6%83%B3%E5%8A%A0%E5%85%A5.png?raw=true)

#### Distance Function

- Kademlia Distance Function : XOR
- the distance between a node and itself is zero
- it is symmetric: "distances" calculaated from A to B and from B to A are the same 
- it follows the triangle inequality

we can't expect each individual node to know about all the other nodes in the network

these facts help us create a virtual ring of nodes so let's see usage of thsi distance function

##### step1

each node has more information about nodes closer to itself and less information about nodes that are further away

each node chooses nodes with 2 to power of n nodes further from itself to ccreate a connection to 

i'm sure you can see that this find request is exponential as each time it is cut to half

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/1.png?raw=true)

now imagine node 0 wants to send data to node 13 , it will choose the closet node to 13 from its reference conncetion or routing table

##### step2

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/2.png?raw=true)

now image node 8 leave network and node 0 wants to send data to ndoe 13

##### step3

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/3.png?raw=true)

also node 0 could pass its own connection information in the request

in fact using Kademlia algorithm we ensure that in n bits guid space there is at most any steps needed to find any other peer from another peer and that's pretty much fast

##### setp4

![](https://github.com/chen4903/selfLearning/blob/main/Bittorrent/imags/4.png?raw=true)

our node will look up for closest notes to the dying node and chooese one to reference 

here we can see node 8 has left the network, and node 0 had node 7 and node 9 in its routing table, and now it should chhose one of them to contact

#### Storing Data in DHT

- Data structure is Key Balue based, like a Map
- Key is usually Hash of value 
- Hash key to GUID space to choose node to store data

Value: "Hello world"

KEY: Hash(value): 255-Hash function with 8 bit size

Node to store: Hash(Key): 15-Hash function with 3 bit size

one important thing to mention is that we can't use our guid space for key range. if we do that, we are saying in a three bit key space we only have eight values. Instead we hash or better to say bound the key to our network size

so imagine our value is "hello world", and our hash function is 8 bit size. and itt says the hash of hello world would be 255. but our network size is three bits so to determine wich node is responsible to store this data we have to bound or hash this key to size of the network. and let's say the result of this hash is 15 in our three bits guid space network. this means that node 15 will be responsible to store the data with key 255 and value "hello world"

#### Commom Practices in DHT

- Copy data to other nodes, so on a node failure there wont be a data loss.
- Move data to closest nodes when node is gracefully shutting down 
- Ask closest nodes for data to store when a new node joins network