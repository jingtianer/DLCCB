NODE=$2
ENODE=$1
AUTHRPCPORT=$(($2+8554))
PORT=$(($2+30308))

echo geth --datadir node$NODE --port $PORT --bootnodes $ENODE --networkid 12345 --unlock 0x$(cat node$NODE/keystore/UTC--* | awk '{split($0, arr, "\""); print arr[4]}') --password node1/password.txt --authrpc.port $AUTHRPCPORT

geth --datadir node$NODE --port $PORT --bootnodes $ENODE --networkid 12345 --unlock 0x$(cat node$NODE/keystore/UTC--* | awk '{split($0, arr, "\""); print arr[4]}') --password node1/password.txt --authrpc.port $AUTHRPCPORT --mine

