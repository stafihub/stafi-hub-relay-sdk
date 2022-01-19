# stafi-hub-relay-sdk




1，eraBlockNumber
使用block timestamp去决定era

2，支持多个endpoint配置，某个rpc挂了，可以自动切换到其他rpc

3，startBlock回退问题（比如回退到bondReported、activeReported事件等），设计一套机制，直接重启即可，程序自动回退处理，不需要手动修改startBlock