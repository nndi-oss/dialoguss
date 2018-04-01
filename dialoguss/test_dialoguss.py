import unittest
from unittest.mock import Mock
import mock_server
from core import Dialoguss, DialStep, Step, AutomatedSession
from unittest import TestCase
DEFAULT_CHANNEL = "384"

class TestAutomatedSession(TestCase):
    def setUp(self):
        mock_server.test_app.testing = True
        # TODO: run this async?? test_app.run(host="0.0.0.0", port=9000)
        pass
    
    def test_should_run_automated_test(self):
        steps = [
            DialStep(expect="What is your name?"),
            Step(1, "Zikani", "Welcome, Zikani\nChoose an item:\n1. Account detail\n2. Balance\n3. Something else\n# Exit"),
            Step(2, "2", "Your balance is: MK 500"),
        ]
        auto = AutomatedSession(
            url="http://localhost:9000/ussd",
            dial="*123#",
            session_id="12345",
            phone_number="265888123456",
            channel=DEFAULT_CHANNEL
        )

        for s in steps:
            s.send_request = Mock(return_value='CON ' + s.expect)

        auto.steps = steps
        auto.run()

        for s in steps:
            s.send_request.assert_called_once_with({
                "sessionId":"12345",
                "phoneNumber":"265888123456",
                "text":s.text,
                "channel":DEFAULT_CHANNEL
            })

if __name__ == "__main__":
    unittest.main()
