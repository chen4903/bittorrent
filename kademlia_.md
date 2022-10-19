# kademlia算法概述

## topics

1. Elements of Kademlia node
2. Network initialization
3. Kademlia API
4. Storing data in Kademlia nodes

## Elements of Kademlia node

- GUID
- Address of node
- Routing Table
- API to call other nodes

a GUID is unique on the identifier for each node that is in certain bit space.

if our bit space is 4 we get maximum 2 to power of 4 guids which means we have node guid from 0 to 15.

Address o node: it could be basically an ip address and a port or it could be an http address or whatever else.

at last each node has an API to accept imcoming requests.

### Routing Table

- Contains List of Buckets
- Bucket List size is same as network GUID space
- List<Bucket<Node>>

the number of pockets we have is same as our network guide space.

nodes that can go in the ends list must have a differing end bit from the node's id

the first n-1 bits of the candidate id must match those of the node's id

this means that it is very easy to populate the first list as half of the nodes in the network are far away from the candidates. the next list can use only quarter of the nodes

#### step1

the 0 is guid of the node, then our second node which is node 1 wants to join the network

since node 1 could get response from node 0

it adds node 0 to its own loading table note that i'm using the word close

node 0 is a bootstrap node since it was used by node 1 to build its initial routing table 

![](imags\节点1想加入.png)

#### step2

![](imags\节点2想加入.png)

### Distance Function

- Kademlia Distance Function : XOR
- the distance between a node and itself is zero
- it is symmetric: "distances" calculaated from A to B and from B to A are the same 
- it follows the triangle inequality

we can't expect each individual node to know about all the other nodes in the network

these facts help us create a virtual ring of nodes so let's see usage of thsi distance function

#### step1

each node has more information about nodes closer to itself and less information about nodes that are further away

each node chooses nodes with 2 to power of n nodes further from itself to ccreate a connection to 

i'm sure you can see that this find request is exponential as each time it is cut to half

![](imags\1.png)

now imagine node 0 wants to send data to node 13 , it will choose the closet node to 13 from its reference conncetion or routing table

#### step2

![](D:\桌面文件\Bittorrent\imags\2.png)

now image node 8 leave network and node 0 wants to send data to ndoe 13

#### step3

![](imags\3.png)

also node 0 could pass its own connection information in the request

in fact using Kademlia algorithm we ensure that in n bits guid space there is at most any steps needed to find any other peer from another peer and that's pretty much fast

#### setp4

![](imags\4.png)

our node will look up for closest notes to the dying node and chooese one to reference 

here we can see node 8 has left the network, and node 0 had node 7 and node 9 in its routing table, and now it should chhose one of them to contact

### Storing Data in DHT

- Data structure is Key Balue based, like a Map
- Key is usually Hash of value 
- Hash key to GUID space to choose node to store data

Value: "Hello world"

KEY: Hash(value): 255-Hash function with 8 bit size

Node to store: Hash(Key): 15-Hash function with 3 bit size

one important thing to mention is that we can't use our guid space for key range. if we do that, we are saying in a three bit key space we only have eight values. Instead we hash or better to say bound the key to our network size

so imagine our value is "hello world", and our hash function is 8 bit size. and itt says the hash of hello world would be 255. but our network size is three bits so to determine wich node is responsible to store this data we have to bound or hash this key to size of the network. and let's say the result of this hash is 15 in our three bits guid space network. this means that node 15 will be responsible to store the data with key 255 and value "hello world"

### Commom Practices in DHT

- Copy data to other nodes, so on a node failure there wont be a data loss.
- Move data to closest nodes when node is gracefully shutting down 
- Ask closest nodes for data to store when a new node joins network