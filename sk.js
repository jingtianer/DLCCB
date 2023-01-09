var keythereum = require("keythereum");
 
var datadir = "/home/tt/eth/net/node2/";
var address= "e43b98ac32beb344c94b15b9af5b46674d6c3e6d";//要小写
const password = "1234567890";
var keyObject = keythereum.importFromFile(address, datadir);
var privateKey = keythereum.recover(password, keyObject);
console.log(privateKey.toString('hex'));
