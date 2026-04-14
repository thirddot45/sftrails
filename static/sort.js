(function () {
    var LOC_COOKIE = "sft_loc";
    var LABEL_COOKIE = "sft_loc_label";
    var MAX_AGE = 60 * 60 * 24; // 1 day

    function setCookie(name, value, maxAge) {
        document.cookie =
            name + "=" + value +
            "; path=/" +
            "; max-age=" + maxAge +
            "; samesite=lax";
    }

    function clearCookie(name) {
        document.cookie = name + "=; path=/; max-age=0";
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
            function (err) {
                alert("Could not get your location: " + err.message);
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
                if (!resp.ok) throw new Error("ZIP not found");
                return resp.json();
            })
            .then(function (data) {
                if (!data.places || !data.places[0]) {
                    throw new Error("ZIP has no location data");
                }
                var place = data.places[0];
                var lat = parseFloat(place.latitude);
                var lng = parseFloat(place.longitude);
                if (isNaN(lat) || isNaN(lng)) {
                    throw new Error("Invalid coordinates from ZIP lookup");
                }
                saveLocation(lat, lng, "ZIP " + zip);
            })
            .catch(function (err) {
                alert("Could not look up ZIP " + zip + ": " + err.message);
            });
        return false;
    };

    window.sftClearLocation = function () {
        clearCookie(LOC_COOKIE);
        clearCookie(LABEL_COOKIE);
        window.location.reload();
    };
})();
