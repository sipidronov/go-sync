# go-sync
Simple client &amp; server implementation for syncing file changes between hosts. As an optimization in-memory index is kept to check if file chunk has changed or not so only real changes sent over the network. Chunk size is hard-coded.

# Dependencies
* zmq (because I like zmq)
* fsnotify

```
brew install zeromq
go get "github.com/howeyc/fsnotify"
go get "github.com/pebbe/zmq4"
```

# Usage
```
./client/client -h
  -h	print help and exit
  -initial-sync
    	perform initial files sync (default true)
  -localpath string
    	Local directory path to track changes
  -optimize
    	do not send un-changed chunks (default true)
  -remote string
    	zmq-like server address to send changes to

./client -localpath '/mnt/test-src' -remote 'tcp://your.example.com:9090'

./server -h
  -endpoint string
    	zmq-like endpoint to listen on
  -h	print help and exit
  -localpath string
    	Local directory path write changes to (default "/tmp")

./server -endpoint tcp://*:9090 -localpath /storage/test-dst
```

# Limitations
* Attributes & extended attributes changes are not synced
* Changed are watched only for files in the directory specified. No recurisve watch for now
* Only file changes (create\modify\delete) expected. Syncing directory changes is not tested and should explode the app.
* Tested only for MacOS and CentOS (should work on other linux distors though)
