mkdir node1 node2
geth --datadir node1 account new < password.txt
geth --datadir node2 account new < password.txt

sed -i "5i$(cat node1/keystore/UTC--* | awk '{split($0, arr, "\""); print arr[4]}')" puppeth.txt 
sed -i "5i$(cat node2/keystore/UTC--* | awk '{split($0, arr, "\""); print arr[4]}')" puppeth.txt 
puppeth < puppeth.txt
sed -i "5d" puppeth.txt
sed -i "5d" puppeth.txt

geth init --datadir node1 tianer.json
geth init --datadir node2 tianer.json

cat password.txt | head -n 1 | tee node1/password.txt
cat password.txt | head -n 1 | tee node2/password.txt

bootnode -genkey boot.key
bootnode -nodekey boot.key -addr :30305

