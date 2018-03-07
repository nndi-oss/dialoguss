DialogUSS
=========

![logo](./logo-small.png)

`dialoguss` is a cli tool to test USSD applications that are implemented
as HTTP services (particularly those implemented on [AfricasTalking's](https://africastalking.com/) 
service or similar).

It allows you to simulate a session and interact with that session
via a simple CLI based interface. USSD "sessions" for your application
can also be tested via automated tests

> NOTE: This is still an early work-in-progress concept

## Usage

### Interactive Dialogue

```yaml
# app.yaml
url: http://localhost:9000/ussd
dial: *123*1234#
phoneNumber: 
```

```sh
$ dialoguss --interactive -f app.yaml
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

### Automated Dialogue

```yaml
#app.yaml
url: http://localhost:9000/ussd
dial: "*1234*1234#"
sessions:
  - id: 12345678910
    phoneNumber: 265888123456
    description: "Should return a balance of 500 for Zikani"
    steps:
      - text: "*123*1234#"
        expect: "What is your name?"
      - text: "name: Zikani"
        expect: "Welcome, Zikani\nChoose an item:\n1. Account detail\n2. Balance\n3. Something else\n# Exit"
      - text: "3"
        expect: "Your balance is: MK 500"
```

```sh
$ dialoguss -f app.yaml
Running tests for session: 12345678910
...
All tests successful
```

---

Copyright (c) 2018, NNDI