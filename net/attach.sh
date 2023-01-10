NODE=$1
echo geth attach node$NODE/geth.ipc

geth attach node$NODE/geth.ipc  < miner.txt