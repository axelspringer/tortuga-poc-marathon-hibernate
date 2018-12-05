(function (ns, undefined) {

    function getStateStatusText(obj) {
        let res = ""
        switch (obj.state) {
            case "hibernate":
                res = obj.host + " is hibernating. Initialize wakeup ...";
                break;
            case "run":
                res = obj.host + " is up and running. Redirecting ...";
                break;
            case "dozeoff":
                res = "Pssst. " + obj.host + " is about to doze off. Preparing hibernation ...";
                break;
            case "wakeup":
                res = obj.host + " is waking up. Wait for readiness ...";
                break;
        }
        return res;
    }

    ns.state = {

    };

    ns.update = function () {
        let xhr = new XMLHttpRequest();
        xhr.onload = function() {
            if (xhr.status !== 200) {
                console.log('Request failed.  Returned status of ' + xhr.status);
                return;
            }

            const hostState = JSON.parse(xhr.responseText);
            let stateText = getStateStatusText(hostState);

            const stateElement = document.getElementById('host-state');
            if (stateElement != undefined) {
                stateElement.innerHTML = stateText;
            }
        };
        xhr.open('GET', 'api/state?host=' + HiberthonHost);
        xhr.send();
    }

    ns.start = function () {
        (function intervalAction() {
            ns.update();
            setTimeout(intervalAction, 2000);
        })();
    }

    ns.update();
    ns.start();

})(window.hiberthon = window.hiberthon || {});