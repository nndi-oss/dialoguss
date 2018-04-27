# -*- coding: utf-8 -*-
"""
dialoguss.core
==============

"""
import logging
import os
import os.path
import re
import sys
import yaml
import requests

from abc import ABCMeta, abstractmethod
from argparse import ArgumentParser

LOGGER = logging.getLogger(__name__)

class SessionCollector(object):
    """Collects information about a session, such as no of requests etc.."""
    def __init__(self, session):
        self.session = session

class Step(object):
    def __init__(self, step_no, text, expect, session=None):
        self.step_no = step_no
        self.text = text
        self.expect = expect
        self.session = session
        self.is_last = True
    
    def send_request(self, data):
        """Sends a request to the http service
        :param: data A dict containing `sessionId`, `phoneNumber`, `text` and `channel` keys
        :return: string or None
        """
        res = requests.post(self.session.url, data)
        response_text = str(res.text)

        if res.status_code not in (200, 201):
            LOGGER.debug('RESPONSE ERROR (%s) : Got an error: %s', res.status_code, response_text)
            response_text = None
        return response_text

    def execute(self, step_input=None):
        """Executes a step and returns the result of the request

        May return an empty string ("") upon failure
        """
        LOGGER.debug("Processing step no: %s", self.step_no)
        text = step_input
        if step_input is None:
            text = ""
        data = {
            'sessionId': self.session.session_id,
            'phoneNumber': self.session.phone_number,
            'text' : text,
            'channel': self.session.channel
        }
        response_text = self.send_request(data)
        if re.search(r'^CON\s?', response_text) is not None:
            # strip out the CONTINUE
            response_text = response_text.replace("CON ", "")
            response_text = response_text.rstrip()
            self.is_last = False
        elif re.search(r'^END\s?', response_text) is not None:
            response_text = response_text.replace("END", "")
            response_text = response_text.rstrip()
            self.is_last = True
        return response_text

class DialStep(Step):
    """DialStep is the first step in the session, dials the USSD service"""
    def __init__(self, expect, text="", session=None):
        super().__init__(0, text, expect, session)

class Session(metaclass=ABCMeta):
    def __init__(self, **kwargs):
        self.url = kwargs['url']
        self.phone_number = kwargs['phone_number']
        self.session_id = kwargs['session_id']
        self.channel = kwargs['channel']
        self.collector = SessionCollector(self)
        self.steps = []

    def add_step(self, step):
        """Add a step for this session"""
        self.steps.append(step)

    @abstractmethod
    def run(self):
        pass

class InteractiveSession(Session):
    """InteractiveSession runs an interactive `USSD` session via the CLI"""
    def run(self):
        step_no = 0
        response_text = DialStep("", "", self).execute()
        sys.stdout.write(response_text + '\n')
        while response_text is not None:
            step_no += 1
            step_input = input("> ")
            a_step = Step(step_no, step_input, "", self)
            response_text = a_step.execute(step_input)
            sys.stdout.write(response_text + '\n')
            if a_step.is_last:
                response_text = None

class AutomatedSession(Session):
    """AutomatedSession runs an automated session that contains pre-defined
    steps (and their expectations)
    """
    def run(self):
        sys.stdout.write("Running tests for session: {}\n".format(self.session_id))
        had_error = False
        for step in self.steps:
            step.session = self
            if isinstance(step, DialStep):
                result = step.execute()
            else:
                result = step.execute(step.text)
            if result != step.expect:
                sys.stderr.write(
                    "StepAssertionError:\n\tExpected={}\n\tGot={}\n".format(step.expect, result))

        if not had_error:
            sys.stdout.write("All tests successful tests for session: {}\n".format(self.session_id))

class Dialoguss:
    """Dialoguss is an application that can have one or more pseudo-ussd sessions"""
    def __init__(self, yamlCfg, is_interactive=False):
        self.config = yamlCfg
        self.is_interactive = is_interactive
        self.session_url = None
        self.dial = None
        self.sessions = []

    def run(self):
        """Runs the main dialoguss application"""
        if self.is_interactive:
            with open(self.config) as f:
                yaml_cfg = yaml.load(f)

            session = InteractiveSession(
                # TODO: Generate a random session id here
                session_id="random_session_id",
                phone_number=yaml_cfg['phoneNumber'],
                channel=yaml_cfg['dial'],
                url=yaml_cfg['url']
            )

            self.sessions.append(session)
            session.run()
        else:
            self.load_sessions()
            for session in self.sessions:
                session.run()

    def load_sessions(self):
        """Loads the sessions for this application"""
        with open(self.config) as f:
            yaml_cfg = yaml.load(f)

        self.session_url = yaml_cfg['url']
        self.dial = yaml_cfg['dial']

        if 'sessions' in yaml_cfg:
            for s in yaml_cfg['sessions']:
                session = AutomatedSession(
                    session_id=s['id'],
                    phone_number=s['phoneNumber'],
                    channel=self.dial,
                    url=self.session_url
                )

                first_step = True
                for i, step in enumerate(s['steps']):
                    if first_step:
                        # session.add_step(DialStep(step.text, step.expect))
                        session.add_step(DialStep(step['expect']))
                        first_step = False
                        continue
                    session.add_step(Step(i, step['text'], step['expect']))

                self.sessions.append(session)
        
def main():
    """Entry point for the CLI program"""
    parser = ArgumentParser(prog="dialoguss")
    parser.add_argument("-i", "--interactive", const='interactive', action='store_const', default=False)
    parser.add_argument("-f", "--file", default="dialoguss.yaml")
    args = parser.parse_args()
    dialoguss_app = Dialoguss(args.file, args.interactive)
    dialoguss_app.run()

if __name__ == "__main__":
    main()
