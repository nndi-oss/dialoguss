from flask import Flask, request, g

test_app = Flask(__name__)
test_app.debug = True

@test_app.route("/ussd", methods = [ 'POST' ])
def test_ussd_handler():
    """Test handler"""
    session_id = request.form['sessionId']
    
    text = request.form['text']

    if text == "":
        return """What is your name?
CONTINUE"""

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