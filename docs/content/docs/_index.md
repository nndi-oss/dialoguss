---
title: 'Overview'
date: 2018-11-28T15:14:39+10:00
weight: 1
---

`dialoguss` is useful for testing your USSD applications during development.

It can be used in two ways:
1. You can simulate a session and interact with that session via a simple CLI based interface or,
2. You can describe the steps required for a session and automate the session, i.e. automated testing

## Usage

`dialoguss` requires a YAML file to run. The file describes one application
and has to contain atleast `url` to the application, `dial` the USSD shortcode for your
app and a `phoneNumber` to use for the session tests.

For the automated dialogue tests you are required to define `steps` which
describe the sequence of steps for one USSD session. Steps define the text
to send to the USSD application and the expected output after sending that 
text.
