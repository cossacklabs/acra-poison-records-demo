# What is this?

Acra Poison Records Demo illustrates how to use intrusion detection functionality of [Acra data protection suite](https://github.com/cossacklabs/acra). For intrusion detection Acra uses poison records, also known as honey tokens. This demo shows how to setup, configure and use intrusion detection of Acra.

# How poison records work?

Poison records are records specifically designed to sit quietly in the database and not be queried by legitimate users under normal circumstances. They looks like any other encrypted records, and it's impossible to distinguish them from "normal data". Technically speaking, poison records is data (binary or strings, int, or whatever suits your database design), placed in particular tables / columns / cells.

However, poison records will only be included in the outputs of suspicious requests from malicious applications that read more data than they should, i.e. using `SELECT *` requests. The sole purpose of these requests is that when an unauthorised leakage occurs, poison records will be present in database response and detected by AcraServer. AcraServer will inform user (system administrator) of untypical behaviour, and can block suspricious request.


Read blog posts:

- [Explain Like I’m Five: Poison Records (Honeypots for Database Tables)](https://hackernoon.com/poison-records-acra-eli5-d78250ef94f)
- [Poison Records In Acra – Database Honeypots For Intrusion Detection](https://www.cossacklabs.com/blog/acra-poison-records.html)
- [Acra docs](https://docs.cossacklabs.com/pages/intrusion-detection/) on intrusion detection 


# How to run the example project

## Installation

1. Use docker-compose command to set up and run the whole infrastructure:

```bash
docker-compose -f docker-infrastructure.yml up
```

This will deploy PostgreSQL database and AcraServer in [transparent mode](https://github.com/cossacklabs/acra#integrating-server-side-encryption-using-acraserver-in-transparent-proxy-mode) of operations.

2. Let's check that those containers are running:

```bash
docker ps -a
```

You should see two containers up and running, and another two in "exited" state (`acra-keymaker` and `acra-poisonrecordmaker`). These containers were used to generate encryption keys for data and poison records themselves. They finished their mission and stopped.

```
CONTAINER ID        IMAGE                                       COMMAND                  CREATED             STATUS                     PORTS                              NAMES
dfcc0e58e111        cossacklabs/acra-server:latest              "/acra-server --conf…"   2 minutes ago       Up 2 minutes               9090/tcp, 0.0.0.0:9393->9393/tcp   acra-poison-records-demo_acra-server_1
2601ddf7fb7b        postgres:11                                 "docker-entrypoint.s…"   2 minutes ago       Up 2 minutes               0.0.0.0:5432->5432/tcp             acra-poison-records-demo_postgresql_1
9726f9355f56        cossacklabs/acra-keymaker:latest            "/acra-keymaker --cl…"   2 minutes ago       Exited (0) 2 minutes ago                                      acra-poison-records-demo_acra-keymaker_server_1
ac0ca175f5be        cossacklabs/acra-poisonrecordmaker:latest   "./acra-poisonrecord…"   2 minutes ago       Exited (0) 2 minutes ago                                      acra-poison-records-demo_acra-poisonrecordmaker_1

```

## Run demo app

Install dependencies and run demo application from repository folder. The demo application is [very simple](https://github.com/cossacklabs/acra-poison-records-demo/blob/master/demo/demo.go), it works as database client application: connects to the database, creates test table, add some encrypted data, add poison records, reads data using `SELECT` query.

```bash
go get gopkg.in/alecthomas/kingpin.v2
go run demo/demo.go --create
```

If no errors occurred, you should see log that table was created:

```
INFO[0000] Table has been successfully created           source="demo.go:65"
```

### Fill in database table with data

Insert some data into table by running client application again, for example, here we add 10 rows:

```bash
go run demo/demo.go --insert 10
```

If no errors, you should see:

```
INFO[0000] Insert has been successful                    source="demo.go:116"
```

Client application adds some random data to the database, but AcraServer sits transparently between app and database, and encrypts all the data before storing in the database.

### Read the database table (aka steal all data)

Let's check that we can read data from the table. Run client application with `--select` command.

```bash
go run demo/demo.go --select
```

If no errors, you should see:

```
INFO[0000] Select has been successful                    source="demo.go:151"
```

Basically, we just downloaded all the content of the table, if we were attackers, we steal all the data successfully. As attackers we could use some SQL injection to perform `SELECT *` query.

### Add poison records to prevent leak

Now we will add poison record to the table to detect attack. Get the value of poison record data from the logs of exited `acra-poisonrecordmaker` container and then insert it into table:

```bash
docker logs acra-poison-records-demo_acra-poisonrecordmaker_1
```

If no errors, you should see base64 encoded value of poison record, it looks like encrypted data that we already have in the database (or like a garbage):

```
IiIiIiIiIiJVRUMyAAAALWSWDMcDH/+0AgCR2bsCZZW47bPtG+WtSD6Riq1PX/NxL1pCpeUgJwQmVAAAAAABAUAMAAAAEAAAACAAAABQeXSzlAcOIYtObhgHLTzGdCKFoEcoBJdtSjmxRtbTZplrFMQMTz15Ieww2FRBbSFN8sH0+pRmtjVxTEWEAAAAAAAAAAABAUAMAAAAEAAAAFgAAAB8UwNKO/MhI0ECetlJfELaqao/L1/WpvrEpGkol2h4MJIl4Mjo2CfEoAICOcJcbfeHPcKCCTtnUFgRhA4b0998U0j5bqBmmFvANHK0mPJMS37xWeLErxUtH/LgJ6ZdDYGg2/TkfS1+cxR/MLuJ93Nkrlf9VQ==
```

Copy poison record from your log of your `acra-poisonrecordmaker` and insert it to the database table:

```bash
go run demo/demo.go --insert_poison IiIiIiIiIiJVRUMyAAAALWSWDMcDH/+0AgCR2bsCZZW47bPtG+WtSD6Riq1PX/NxL1pCpeUgJwQmVAAAAAABAUAMAAAAEAAAACAAAABQeXSzlAcOIYtObhgHLTzGdCKFoEcoBJdtSjmxRtbTZplrFMQMTz15Ieww2FRBbSFN8sH0+pRmtjVxTEWEAAAAAAAAAAABAUAMAAAAEAAAAFgAAAB8UwNKO/MhI0ECetlJfELaqao/L1/WpvrEpGkol2h4MJIl4Mjo2CfEoAICOcJcbfeHPcKCCTtnUFgRhA4b0998U0j5bqBmmFvANHK0mPJMS37xWeLErxUtH/LgJ6ZdDYGg2/TkfS1+cxR/MLuJ93Nkrlf9VQ==
```

If no errors, you should see base64 encoded value of poison record:

```
INFO[0000] Poison record insert has been successful      source="demo.go:136"
```

### Try to steal data again

Now we are protected from malicious `SELECT *` queries. Try to read all data again:

```bash
go run demo/demo.go --select
```

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

It means that AcraServer detected poison record and stopped working (shut down itself). Note, that you can setup AcraServer to different behaviour when it detects poison record:

- perform a shut-down (useful for very critical data, but AcraServer will be down until you restart it)
- run a script (you can tell AcraServer to run a script after detecting poison record, for example, to send alerts to system administrators and SIEMs)
- perform a shut-down and run a script

# Further steps

Let us know if you have any questions by dropping an email to [dev@cossacklabs.com](mailto:dev@cossacklabs.com).

1. [cossacklabs/acra](https://github.com/cossacklabs/acra) – the main Acra repository contains tons of examples and documentation.
2. Check dozens of Acra-based applications and configuration examples in [Acra Engineering Demo](https://github.com/cossacklabs/acra-engineering-demo/) repository.
3. [Acra Live Demo](https://www.cossacklabs.com/acra/#acralivedemo) – is a web-based demo of a typical web-infrastructure protected by Acra and deployed on our servers for your convenience. It illustrates the other features of Acra, i.e. SQL firewall, intrusion detection, database rollback, and so on.

