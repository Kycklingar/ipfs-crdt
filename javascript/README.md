Try it out http://localhost:8080/ipfs/zdj7Wf5Ku7v5E7gXTJgZRK85SA1vuBaTh6qqQH1Ht4ATgC89h

This is a quickly hacked together JS version of https://github.com/kycklingar/ipfs-crdt.

Requirements:

    go-ipfs daemon with --enable-pubsub-experiment and ipfs API access on localhost:5001
    enable API access with 
    "ipfs config --json API.HTTPHeaders.Access-Control-Allow-Methods '["GET","POST"]'"
    "ipfs config --json API.HTTPHeaders.Access-Control-Allow-Origin '["*"]'"
