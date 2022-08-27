---
title: 'Interactive Dialogue'
date: 2019-02-11T19:27:37+10:00
draft: false
weight: 3
---

```yaml
# app.yaml
url: http://localhost:9000/ussd
dial: *123*1234#
phoneNumber: 265888123456
```

```sh
$ dialoguss -i -f app.yaml
Sending *123*1234# to <app>
USSD Response:
What is your name?
> name: Zikani

Hello Zikani, choose an item:
1. Account detail
2. Balance
3. Something else
# Exit
> 2

Your balance is: MK 500
> ok
```


## Why should I use this?

Glad you asked! Well, mostly this tool will help you reduce costs 
related to testing your USSD applications.
The current approach for testing your applications could be to upload 
the code to your server, pull out your phone and dial the USSD service 
code linked to your application. 

_That's too much work and costs you time and monies!_

You should use this if you'd like to test your application before deploying 
it to production.

---

Copyright (c) 2018 - 2020, NNDI