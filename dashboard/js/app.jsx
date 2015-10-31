var Dashboard = React.createClass({
    getInitialState: function() {
        return {};
    },

    componentDidMount: function() {
        this.reload();
    },

    reload: function() {
        getURL("http://127.0.0.1:6464/stats.json", {}, function(resp) {
            var newState = {};
            var decode = function(point) {
                return {
                    timestamp: point[0],
                    processed: point[1],
                    errors:    point[2],
                    min:       point[3],
                    p25:       point[4],
                    mean:      point[5],
                    median:    point[6],
                    p75:       point[7],
                    max:       point[8],
                }
            };
            for (name in resp) {
                newState[name] = resp[name].map(decode);
            }
            this.setState(newState);
            setTimeout(this.reload, 3000);
        }.bind(this));
    },

    renderDaemons: function() {
        var daemons = [];
        for (name in this.state) {
            daemons.push(<Daemon name={name} key={name} points={this.state[name]} />);
        }

        return daemons;
    },

    render: function() {
        return (
            <div className="dashboard">
                <Header />
                <div className="daemons">{this.renderDaemons()}</div>
            </div>
        );
    }
});

var Header = React.createClass({
    render: function() {
        return (
            <div className="header">
                <div className="col-name">Daemon</div>
                <div className="col-processed">Processed</div>
                <div className="col-errors">Errors</div>
                <div className="col-min">Min</div>
                <div className="col-median">Median</div>
                <div className="col-max">Max</div>
            </div>
        );
    }
});

var StatsRow = React.createClass({
    render: function() {
        var value = this.props.value;
        return (
            <div className="stats">
                <div className="col-name">{this.props.name}</div>
                <div className="col-processed">{value.processed}</div>
                <div className="col-errors">{value.errors}</div>
                <div className="col-min">{formatDuration(value.min)}</div>
                <div className="col-median">{formatDuration(value.median)}</div>
                <div className="col-max">{formatDuration(value.max)}</div>
            </div>
        );
    }
});

var Daemon = React.createClass({
    render: function() {
        var last = this.props.points[this.props.points.length - 1];
        return (
            <div className="daemon">
                <StatsRow name={this.props.name} value={last} />
                <div className="left-block">
                    <LineChart points={this.props.points} />
                </div>
                <BoxPlot points={this.props.points} />
            </div>
        );
    }
});

var BoxPlot = React.createClass({
    render: function(){
        var points = this.props.points,
            maxHeight = 140,
            padding = 5;

        var min, max;
        points.map(function(point) {
            if (min === undefined || point.min < min) {
                min = point.min;
            }
            if (max === undefined || point.max > max) {
                max = point.max;
            }
        });

        var renderBox = function(point, i) {
            var relativeY = function(val) {
                return maxHeight - Math.round((val-min)/(max-min) * maxHeight) + padding;
            };

            var width = 10;
            var padding = 5;

            var x1 = (width + padding) * i + padding;
            var x2 = x1 + width;
            var minY = relativeY(point.min);
            var p25Y = relativeY(point.p25);
            var medianY = relativeY(point.median);
            var p75Y = relativeY(point.p75);
            var maxY = relativeY(point.max);

            return (
                <g key={i}>
                    <line key="max"
                        x1={x1+2}
                        x2={x2-2}
                        y1={maxY}
                        y2={maxY}
                        strokeWidth={1}
                        style={{stroke: "#aaa"}} />
                    <line key="max-bar"
                        x1={x1+width/2}
                        x2={x1+width/2}
                        y1={maxY}
                        y2={p75Y}
                        strokeDasharray="3,1"
                        strokeWidth={1}
                        style={{stroke: "#ccc"}} />
                    <line key="min-bar"
                        x1={x1+width/2}
                        x2={x1+width/2}
                        y1={minY}
                        y2={p25Y}
                        strokeDasharray="3,1"
                        strokeWidth={1}
                        style={{stroke: "#ccc"}} />
                    <rect key="iqr"
                        x={x1}
                        y={p75Y}
                        width={width}
                        height={p25Y - p75Y}
                        strokeWidth={1}
                        style={{fill: "#f0f0f0", stroke: "#888"}} />
                    <line key="median"
                        x1={x1}
                        x2={x2}
                        y1={medianY}
                        y2={medianY}
                        strokeWidth={2}
                        style={{stroke: "#444"}} />
                    <line key="min"
                        x1={x1+2}
                        x2={x2-2}
                        y1={minY}
                        y2={minY}
                        strokeWidth={1}
                        style={{stroke: "#aaa"}} />
                </g>
            );
        };
        return (
            <div className="boxplot">
                <svg width="455" height="150">
                    {this.props.points.map(renderBox)}
                </svg>
            </div>
        );
    }
});

var LineChart = React.createClass({
    render: function() {
        var points = this.props.points,
            maxHeight = 90,
            padding = 5,
            colors = {processed: "#46f", errors: "#f64"};

        var min = 0, max;
        points.map(function(point) {
            if (max === undefined || point.processed > max) {
                max = point.processed;
            }
        });

        var makePath = function(points, key) {
            if (max === 0) {
                return;
            }

            var path = points.map(function(point, i) {
                var val = point[key];
                var width = 15;
                var x = i * width;
                var y = maxHeight - Math.round((val-min)/(max-min) * maxHeight) + padding;

                if (i === 0) {
                    return "M"+x+","+y;
                } else {
                    return "L"+x+","+y;
                }
            });

            return (
                <path
                    d={path.join(" ")}
                    strokeWidth={2}
                    style={{stroke: colors[key], fill: "transparent"}} />
            );
        };

        return (
            <div className="linechart">
                <svg width="300" height="100">
                    {makePath(points, "processed")}
                    {makePath(points, "errors")}
                </svg>
            </div>
        );
    }
});

ReactDOM.render(<Dashboard />, document.getElementById("app"));
