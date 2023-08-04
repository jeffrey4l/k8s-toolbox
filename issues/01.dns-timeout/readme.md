* https://monkeywie.cn/2019/12/10/k8s-dns-lookup-timeout/
* single-request-reopen (glibc>=2.9)
发送 A 类型请求和 AAAA 类型请求使用不同的源端口。这样两个请求在 conntrack 表中不占用同一个表项，从而避免冲突。
* single-request (glibc>=2.10)
避免并发，改为串行发送 A 类型和 AAAA 类型请求，没有了并发，从而也避免了冲突。

```
# options use-vc

./dns_query  -host mariadb-0.mariadb.mariadb.svc -c 100 -d 10 -l 3000
lookup mariadb-0.mariadb.mariadb.svc on 169.254.25.10:53: dial tcp 169.254.25.10:53: connect: cannot assign requested address
request count：4676
error count：1
request time：min(8ms) max(2318ms) avg(192ms) timeout(0n)


# options single-request-reopen

./dns_query  -host mariadb-0.mariadb.mariadb.svc -c 100 -d 10 -l 3000
request count：19397
error count：0
request time：min(2ms) max(191ms) avg(50ms) timeout(0n)
```

```bash
## 173
./dns_query  -host mariadb-0.mariadb.mariadb.svc -c 100 -d 10 -l 3000
lookup mariadb-0.mariadb.mariadb.svc on 169.254.25.10:53: dial tcp 169.254.25.10:53: connect: cannot assign requested address
request count：4674
error count：1
request time：min(5ms) max(3465ms) avg(188ms) timeout(13n)

## mec41, 概率稍微小
./a -host loki-0.loki-headless.loki.svc -d 20  -l 3000
request count：3735
error count：0
request time：min(1ms) max(6398ms) avg(516ms) timeout(23n)


## mec21
./dns_query -host loki-0.loki-headless.loki.svc -d 20  -l 3000
request count：103727
error count：0
request time：min(1ms) max(5059ms) avg(18ms) timeout(58n)
```

# REF

* https://monkeywie.cn/2019/12/10/k8s-dns-lookup-timeout/
