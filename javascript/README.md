This is a quickly hacked together JS version of https://github.com/kycklingar/ipfs-crdt.
Requirements:
    go-ipfs daemon with --experimental-pubsub-enabled and ipfs API access on localhost:5001
    enable API access with "ipfs config --json API.HTTPHeaders.Access-Control-Allow-Methods '["GET","POST"]'" and "ipfs config --json API.HTTPHeaders.Access-Control-Allow-Origin '["*"]'"