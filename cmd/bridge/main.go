package main

/*
CENTRAL SERVER <-> BRIDGE SERVER
- central server receives HTTP request. Sends that request to the bridge server.
- There needs to be a GRPC stream between central and bridge. BUT, central becomes the "host"
- Only need GRPC connection is needed
- central server acts as GRPC server, and the bridges dial in
- M connections, M=# bridges/projects that are active

BRIDGE SERVER <-> CLIENT
- this connects to multiple N clients
- can have N GRPC connections
- bridge acts as the GRPC server, and then clients dial in
- N connections

There is a single STREAM method. This should be shared between the BRIDGE and the CENTRAL servers.
How does the HTTP method get back the response in Gin?
- HTTP request comes in to central server
- That is in its own function. There is already a stream between. So it just sends through the stream
- Then, the bridge server eventually returns. There is a separate receiver thread/goroutine
- This separates goroutine will probably add to a buffer/queue, depending on the request id (demultiplexing)
- WAIT NO; this is just the same as using a channel;
- basically, after we send the HTTP request from central to bridge. The HTTP method now waits upon a channel between the RECEIVER goroutine
- This channel will be size 1, since request id is unique.
- Then, we can send result back as a response

How does the flow look like between BRIDGE and CLIENT?
- Bridge receives a request object from central server. This is a receiver goroutine
- But, it needs to get to the client grpc sender now. it can directly just call the stream send method though. Now, it waits on a channel. This goroutine is identified by requestID
- BUT, it should call this in a NEW GOROUTINE, bc there is only 1 receiver goroutine that is listening for CENTRAL requests.
- BUT, there is again, a receiver goroutine from the client. It sends to a channel, which reactivates the goroutine that sent originally.
- Back to the goroutine that's waiting on the channel, this now sends it back?
- Technically, there is no need for channel. Can just send from wherever we are, bc it's not like the response needs to be retunred from same function
- BUT, this might be better design/isolation, and shared logic perhaps

What can be shared?
- Receiver Goroutine: central is listening for RESPONSE, and then sends on a channel based on ID (map can store)
	- BRIDGE is listening for RESPONSE, then also sends to the goroutine for requestID.
	- these goroutine implementations will be different, but the receiver logic should be the same


We need 1 receiver goroutine BECAUSE we are using ONLY 1 STREAM RPC
- in this usecase, they are all sharing the same stream to avoid renegotiating handshake or whatever for every new message
- if I did that, then it would be separate RPC streams, but that = overhead. There is no need to use HTTP/2 like pattern of mixing chunks because we aren't usually sending super large objects like videos
*/
