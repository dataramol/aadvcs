@ECHO OFF
ECHO Executing Testing Script
del .\aadvcs.exe
rmdir .\.aadvcs
make build-cli
.\aadvcs.exe init
.\aadvcs.exe add .\dir\
.\aadvcs.exe commit -m "test commit"