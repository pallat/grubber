How to Build stub binary
========================
1. build stub binary
~~~
./build
mv main stub
~~~
2. remote to **andromeda11 & andromeda12** and kill stub process
~~~
pkill stub
~~~
3. turn back to your local machine and copy a new binary file to **andromeda11 & andromeda12**
~~~
scp stub omropr@andromeda11.tac.co.th:stub/stub
scp stub omropr@andromeda12.tac.co.th:stub/stub
~~~
4. go to [Jenkins: Grubber][1] and click build it

log file
--------
trace.log is on andromeda11 & andromeda12 at `/home/omropr/stub`

How to make new cache
=====================
1. run local stub
``` 
go run main.go -port=:8765
```
2. run local api with proxy
```
go run main/server.go -proxy=http://localhost:8765
```
3. make a request you want the cache to local api
4. the cache are generated at `stub/cache` path
5. you can see `stub/trace.log` to diagnose which file you want
6. copy cached file from `stub/cache` to `stub/saved`
7. commit stub repo and push to remote repo
8. go to [Jenkins: Grubber][1] and build it

[1]: http://10.89.104.33/job/Grubber/ "Grubber"