"""This is a simple mock ussd serve. DO NOT use this as an example of 
what an actual USSD request handling server should be implemented, please.
"""
import os
import re
from flask import Flask, request, session

test_app = Flask(__name__)
test_app.debug = True
test_app.secret_key = os.urandom(5)

@test_app.route("/ussd", methods = [ 'POST' ])
def test_ussd_handler():
    """Test handler"""
    session_id = request.form['sessionId']
    text = request.form['text']
    test_app.logger.debug('Got text: {}'.format(text))

    if len(text) < 1:
        return """What is your name?
CONTINUE"""

    
    if re.match("name\:", text) is not None:
        name = text.replace("name:", "").strip()
        res_ = """Welcome, {}
Choose an item:
1. Account detail
2. Balance
3. Something else
# Exit
CONTINUE
""".format(name)
        return res_

    if text == "1" or text == "1.":
        return """Your account is inactive
CONTINUE
"""

    if text == "2" or text == "2.":
        return """Your balance is: MK 500
CONTINUE
"""

    if text == "3" or text == "3.":
        return """There is nothing else, /(^_^)\\
END
"""
    return "END"

if __name__ == '__main__':
    test_app.run(host="0.0.0.0", port=9000)