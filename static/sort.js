(function () {
    var LOC_COOKIE = "sft_loc";
    var LABEL_COOKIE = "sft_loc_label";
    var MAX_AGE = 60 * 60 * 24; // 1 day

    function setCookie(name, value, maxAge) {
        // Add Secure only on HTTPS so local http://localhost dev still works.
        var secure = location.protocol === "https:" ? "; secure" : "";
        document.cookie =
            name + "=" + value +
            "; path=/" +
            "; max-age=" + maxAge +
            "; samesite=lax" + secure;
    }

    function clearCookie(name) {
        var secure = location.protocol === "https:" ? "; secure" : "";
        document.cookie = name + "=; path=/; max-age=0; samesite=lax" + secure;
    }

    function saveLocation(lat, lng, label) {
        // Round to 2 decimal places (~0.7 mile precision) to avoid
        // storing more precision than we need.
        var roundedLat = Math.round(lat * 100) / 100;
        var roundedLng = Math.round(lng * 100) / 100;
        setCookie(LOC_COOKIE, roundedLat + "," + roundedLng, MAX_AGE);
        setCookie(LABEL_COOKIE, encodeURIComponent(label), MAX_AGE);
        window.location.reload();
    }

    window.sftUseGeolocation = function () {
        if (!navigator.geolocation) {
            alert("Geolocation is not supported in this browser.");
            return;
        }
        navigator.geolocation.getCurrentPosition(
            function (pos) {
                saveLocation(pos.coords.latitude, pos.coords.longitude, "my location");
            },
            function () {
                alert("Could not get your location. Check browser permissions and try again.");
            },
            { enableHighAccuracy: false, timeout: 10000, maximumAge: 600000 }
        );
    };

    window.sftUseZip = function (evt) {
        evt.preventDefault();
        var input = document.getElementById("sft-zip");
        if (!input) return false;
        var zip = input.value.trim();
        if (!/^\d{5}$/.test(zip)) {
            alert("Please enter a 5-digit ZIP code.");
            return false;
        }
        fetch("https://api.zippopotam.us/us/" + zip)
            .then(function (resp) {
                if (!resp.ok) throw new Error("lookup_failed");
                return resp.json();
            })
            .then(function (data) {
                if (!data.places || !data.places[0]) {
                    throw new Error("lookup_failed");
                }
                var place = data.places[0];
                var lat = parseFloat(place.latitude);
                var lng = parseFloat(place.longitude);
                if (isNaN(lat) || isNaN(lng)) {
                    throw new Error("lookup_failed");
                }
                saveLocation(lat, lng, "ZIP " + zip);
            })
            .catch(function () {
                alert("Could not look up ZIP " + zip + ".");
            });
        return false;
    };

    window.sftClearLocation = function () {
        clearCookie(LOC_COOKIE);
        clearCookie(LABEL_COOKIE);
        window.location.reload();
    };
})();
