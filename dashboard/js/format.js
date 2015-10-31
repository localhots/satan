function formatDuration(dur) {
    if (dur < 10) {
        return "" + dur.toFixed(3) + "ms";
    } else if (dur < 100) {
        return "" + dur.toFixed(2) + "ms";
    } else if (dur < 1000) {
        return "" + dur.toFixed(1) + "ms";
    } else if (dur < 10000) {
        return "" + (dur / 1000).toFixed(3) + "s";
    } else if (dur < 60000) {
        return "" + (dur / 1000).toFixed(2) + "s";
    } else {
        dur = Math.ceil(dur / 1000);
        var m = Math.floor(dur / 60);
        var s = dur % 60;
        return "" + m + "m" + s + "s";
    }
}
