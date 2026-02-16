(function() {
    function hash(str) {
        var h = 5381;
        for (var i = 0; i < str.length; i++) {
            h = ((h << 5) + h + str.charCodeAt(i)) & 0xffffffff;
        }
        return (h >>> 0).toString(16);
    }

    function getFingerprint() {
        var parts = [
            navigator.userAgent || "",
            screen.width + "x" + screen.height,
            Intl.DateTimeFormat().resolvedOptions().timeZone || "",
            navigator.language || ""
        ];
        return hash(parts.join("|"));
    }

    function setFingerprint() {
        var fp = getFingerprint();
        var inputs = document.querySelectorAll('input[name="fingerprint"]');
        for (var i = 0; i < inputs.length; i++) {
            inputs[i].value = fp;
        }
    }

    document.addEventListener("DOMContentLoaded", setFingerprint);
    document.addEventListener("htmx:afterSwap", setFingerprint);
})();
