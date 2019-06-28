# acra-poison-records-demo

This project illustrates how to use Acra's intrusion detection functionality (poison records).

Poison records are the records specifically designed and crafted in such a way that they wouldn't
be queried by a user under normal circumstances. Yet poison records will be included in the outputs
of `SELECT *` requests. Upon passing AcraServer, they will inform it of untypical behaviour. The goal
of using poison records is simple — to detect adversaries trying to download full tables / full database
from the application server or trying to run full scans in their injected queries

## How to run the demo

**1)** Use docker-compose command to set up and run the whole infrastructure:

`docker-compose -f docker-infrastructure.yml up`

This will deploy PostgreSQL database and AcraServer in transparent mode of operations.

**2)** Let's check that those containers are running:

`docker ps -a`

You should also see two additional exited containers (`acra-keymaker` and `acra-poisonrecordmaker`)
used for cryptographic keys generation and poison record creation, respectively. They have already done their tasks

```
CONTAINER ID        IMAGE                                       COMMAND                  CREATED             STATUS                     PORTS                              NAMES
dfcc0e58e111        cossacklabs/acra-server:latest              "/acra-server --conf…"   2 minutes ago       Up 2 minutes               9090/tcp, 0.0.0.0:9393->9393/tcp   acra-poison-records-demo_acra-server_1
2601ddf7fb7b        postgres:11                                 "docker-entrypoint.s…"   2 minutes ago       Up 2 minutes               0.0.0.0:5432->5432/tcp             acra-poison-records-demo_postgresql_1
9726f9355f56        cossacklabs/acra-keymaker:latest            "/acra-keymaker --cl…"   2 minutes ago       Exited (0) 2 minutes ago                                      acra-poison-records-demo_acra-keymaker_server_1
ac0ca175f5be        cossacklabs/acra-poisonrecordmaker:latest   "./acra-poisonrecord…"   2 minutes ago       Exited (0) 2 minutes ago                                      acra-poison-records-demo_acra-poisonrecordmaker_1

```

**3)** Run demo application and create table in our database:

`go run demo/demo.go --create`

If no errors, you should see:

```
INFO[0000] Table has been successfully created           source="demo.go:65"
```

**4)** Run demo application and insert some data (for example 10 rows) into the created table:#

`go run demo/demo.go --insert 10`

If no errors, you should see:

```
INFO[0000] Insert has been successful                    source="demo.go:116"
```

Let's check that we can select data:

**4)** Run demo application and select all data from our table:

`go run demo/demo.go --select`

If no errors, you should see:

```
INFO[0000] Select has been successful                    source="demo.go:151"
```

Let's suppose, some SQL injection was used and adversary injected `select *` query to our table with sensitive data.
To prevent this, we should add poison records to our table:

**5)** Get poison record value from the logs of exited `acra-poisonrecordmaker` container and then insert it into table:

`docker logs acra-poison-records-demo_acra-poisonrecordmaker_1`

If no errors, you should see base64 encoded value of poison record:

```
IiIiIiIiIiJVRUMyAAAALWSWDMcDH/+0AgCR2bsCZZW47bPtG+WtSD6Riq1PX/NxL1pCpeUgJwQmVAAAAAABAUAMAAAAEAAAACAAAABQeXSzlAcOIYtObhgHLTzGdCKFoEcoBJdtSjmxRtbTZplrFMQMTz15Ieww2FRBbSFN8sH0+pRmtjVxTEWEAAAAAAAAAAABAUAMAAAAEAAAAFgAAAB8UwNKO/MhI0ECetlJfELaqao/L1/WpvrEpGkol2h4MJIl4Mjo2CfEoAICOcJcbfeHPcKCCTtnUFgRhA4b0998U0j5bqBmmFvANHK0mPJMS37xWeLErxUtH/LgJ6ZdDYGg2/TkfS1+cxR/MLuJ93Nkrlf9VQ==
```

Run:

`go run demo/demo.go --insert_poison IiIiIiIiIiJVRUMyAAAALWSWDMcDH/+0AgCR2bsCZZW47bPtG+WtSD6Riq1PX/NxL1pCpeUgJwQmVAAAAAABAUAMAAAAEAAAACAAAABQeXSzlAcOIYtObhgHLTzGdCKFoEcoBJdtSjmxRtbTZplrFMQMTz15Ieww2FRBbSFN8sH0+pRmtjVxTEWEAAAAAAAAAAABAUAMAAAAEAAAAFgAAAB8UwNKO/MhI0ECetlJfELaqao/L1/WpvrEpGkol2h4MJIl4Mjo2CfEoAICOcJcbfeHPcKCCTtnUFgRhA4b0998U0j5bqBmmFvANHK0mPJMS37xWeLErxUtH/LgJ6ZdDYGg2/TkfS1+cxR/MLuJ93Nkrlf9VQ==`

If no errors, you should see base64 encoded value of poison record:

```
INFO[0000] Poison record insert has been successful      source="demo.go:136"
```

**6)** Now we are protected from malicious `select *` queries. Run:

`go run demo/demo.go --select`

You should see:

```
FATA[0000] read tcp 127.0.0.1:58266->127.0.0.1:9393: read: connection reset by peer  source="demo.go:148"
exit status 1
```

Also, check the console where you run infrastructure. You should see that poison records has been detected by AcraServer:

```
acra-server_1             | time="2019-06-28T12:07:02Z" level=warning msg="Recognized poison record" client_id=poison_records_demo code=587
acra-server_1             | time="2019-06-28T12:07:02Z" level=warning msg="detected poison record, exit"
```

Note, that AcraServer is set to shutdown, but there are 3 variants of behaviour setting:

- perform a shut-down;
- run a script if a poison record is matched in the input stream;
- perform a shut-down and run a script.

Check our blog-posts (https://hackernoon.com/poison-records-acra-eli5-d78250ef94f, https://www.cossacklabs.com/blog/acra-poison-records.html) and documentation (https://docs.cossacklabs.com/pages/intrusion-detection/) for additional information.
