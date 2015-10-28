function formatDuration(dur) {
    if (dur < 10) {
        return "" + dur + "ms";
    } else if (dur < 100) {
        return "" + dur.toFixed(5) + "ms";
    } else if (dur < 1000) {
        return "" + dur.toFixed(4) + "ms";
    } else if (dur < 10000) {
        return "" + (dur / 1000).toFixed(6) + "s";
    } else if (dur < 60000) {
        return "" + (dur / 1000).toFixed(5) + "s";
    } else {
        dur = Math.round(dur / 1000);
        var m = Math.floor(dur / 60);
        var s = dur % 60;
        return "" + m + "m" + s + "s";
    }
}
