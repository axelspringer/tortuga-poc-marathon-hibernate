(function (ns, undefined) {

    function getStateStatusText(obj) {
        
    }

    ns.state = {

    };

    ns.wakeup = function () {
        let xhr = new XMLHttpRequest();
        xhr.open("POST", '/-/api/trigger', true);
        xhr.setRequestHeader("Content-Type", "application/json");

        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4 && xhr.status === 201) {
                console.log('trigger ' + xhr.responseText);
            }
        };

        var data = JSON.stringify([ns.host]);
        xhr.send(data);
    };

    ns.update = function () {
        let xhr = new XMLHttpRequest();
        xhr.onload = function() {
            if (xhr.status !== 200) {
                console.log('Request failed.  Returned status of ' + xhr.status);
                return;
            }

            const hostState = JSON.parse(xhr.responseText);
            let res = ""
            switch (hostState.state) {
                case "hibernate":
                    res = ns.host + " is hibernating. Initialize wakeup ...";
                    ns.wakeup(ns.host);
                    break;
                case "run":
                    res = ns.host + " is up and running. Redirecting ...";
                    setTimeout(function(){ window.location.reload(); }, 2000);
                    break;
                case "dozeoff":
                    res = "Pssst. " + ns.host + " is about to doze off. Preparing hibernation ...";
                    break;
                case "wakeup":
                    res = ns.host + " is waking up. Wait for readiness ...";
                    break;
            }

            const stateElement = document.getElementById('host-state');
            if (stateElement != undefined) {
                stateElement.innerHTML = res;
            }
        };
        xhr.open('GET', '/-/api/state?host=' + btoa(ns.host));
        xhr.send();
    }

    ns.start = function () {
        (function intervalAction() {
            if (ns.host != undefined) {
                ns.update();
            }
            setTimeout(intervalAction, 2000);
        })();
    }

})(window.hiberthon = window.hiberthon || {});

$(document).ready(function () {
    window.hiberthon.referer = window.location.href;
    window.hiberthon.host = window.location.hostname;
    window.hiberthon.update();
    window.hiberthon.start();
});