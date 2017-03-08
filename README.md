How to Build
-----------
~~~
./build
mv main stub
~~~

## remote to **andromeda11 & andromeda12** and kill stub process
`pkill stub`

## turn back to your local machine and copy a new binary file to **andromeda11 & andromeda12**
~~~
scp stub omropr@andromeda11.tac.co.th:stub/stub
scp stub omropr@andromeda12.tac.co.th:stub/stub
~~~

## go to Jenkins: Grubber and click build it

log file
--------
trace.log is on andromeda11 & andromeda12 at /home/omropr/stub