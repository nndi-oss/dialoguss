import unittest
from mock_server import test_app
from unittest import TestCase
DEFAULT_CHANNEL = "384"

class TestAutomatedSession(TestCase):
    def setUp(self):
        # TODO: run this async?? test_app.run(host="0.0.0.0", port=5000)
        pass
    
    def test_should_run_automated_test(self):
        steps = [
            DialStep(expect="What is your name?"),
            Step(1, "Zikani", "Welcome Zikani\nChoose a menu item:"),
            Step(2, "2", "Your balance is: MK 500"),
        ]
        auto = AutomatedSession(
            url="http://localhost:5000/ussd",
            dial="*123#",
            session_id="12345",
            phone_number="265888123456",
            channel=DEFAULT_CHANNEL
        )

        auto.steps = steps
        auto.run()

if __name__ == "__main__":
    unittest.main()
