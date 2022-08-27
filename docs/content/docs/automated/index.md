---
title: 'Automated Dialogue'
date: 2019-02-11T19:27:37+10:00
draft: false
weight: 3
---

```yaml
# app.yml
url: http://localhost:7654
dial: "*1234*1234#"
# 'global' phone number, overriden per session
phoneNumber: 265888123456
sessions:
  - id: 12345678910
    phoneNumber: 265888123456
    description: "Should return a balance of 500 for Zikani"
    steps:
      # The first step is the response after dialing the ussd code
      - expect: "What is your name?"
      - userInput: "Zikani"
        expect: |-
          Welcome, Zikani
          Choose an item:
          1. Account detail
          2. Balance
          3. Something else
          # Exit
      - userInput: "2" 
        expect: "Your balance is: MK 500"
```

```sh
$ dialoguss -f app.yml
All steps in session 12345678910 run successfully
```