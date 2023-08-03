@ECHO OFF
ECHO Executing Testing Script
del .\aadvcs.exe
del /s /q .\.aadvcs
rmdir /s /q .\.aadvcs
make build-cli
.\aadvcs.exe init
.\aadvcs.exe add .\dir\
.\aadvcs.exe commit -m "test commit"