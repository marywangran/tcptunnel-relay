# tcptunnel-relay

HostA--(cloud)[use BBR]--HostM--(cloud)[use CUBIC]--HostB

HostM:relay point
```bash
./relay -h 192.168.10.100 -p 1234
```

HostA:end point
```bash
./edge -h 192.168.10.100 -p 1234
ifconfig edge 172.16.0.1/24 up
iperf -s
```


HostB:end point
```bash
./edge -h 192.168.10.100 -p 1235
ifconfig edge 172.16.0.2/24 up
iperf -c 172.16.0.1 -i 1
```

![image](https://user-images.githubusercontent.com/6219622/124539505-356cba00-de50-11eb-965b-7629ee4dc525.png)
