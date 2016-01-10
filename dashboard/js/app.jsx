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
            <div className="table-row table-header">
                <div className="table-cell col-name">Daemon</div>
                <div className="table-cell col-processed">Processed</div>
                <div className="table-cell col-errors">Errors</div>
                <div className="table-cell col-min">Min</div>
                <div className="table-cell col-median">Median</div>
                <div className="table-cell col-max">Max</div>
            </div>
        );
    }
});

var StatsRow = React.createClass({
    render: function() {
        var value = this.props.value;
        return (
            <div className="table-row">
                <div className="table-cell col-name">{this.props.name}</div>
                <div className="table-cell col-processed">{value.processed}</div>
                <div className="table-cell col-errors">{value.errors}</div>
                <div className="table-cell col-min">{formatDuration(value.min)}</div>
                <div className="table-cell col-median">{formatDuration(value.median)}</div>
                <div className="table-cell col-max">{formatDuration(value.max)}</div>
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
                <LineChart points={this.props.points} />
                <BoxPlot points={this.props.points} />
            </div>
        );
    }
});

var BoxPlot = React.createClass({
    render: function(){
        var points = this.props.points,
            chartWidth = 950,
            chartHeight = 170,
            padLeft = 30,
            padTop = 10,
            padBottom = 20,
            valueInterval = 5,
            maxHeight = chartHeight - padTop - padBottom;

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
                return maxHeight - Math.round((val-min)/(max-min) * maxHeight);
            };

            var boxWidth = 10;

            var x1 = (boxWidth + valueInterval) * i + padLeft + valueInterval;
            var x2 = x1 + boxWidth;
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
                        className="tick" />
                    <line key="max-whisker"
                        x1={x1+boxWidth/2}
                        x2={x1+boxWidth/2}
                        y1={maxY}
                        y2={p75Y}
                        className="whisker" />
                    <line key="min-whisker"
                        x1={x1+boxWidth/2}
                        x2={x1+boxWidth/2}
                        y1={minY}
                        y2={p25Y}
                        className="whisker" />
                    <rect key="iqr"
                        x={x1}
                        y={p75Y}
                        width={boxWidth}
                        height={p25Y - p75Y}
                        className="iqr" />
                    <line key="median"
                        x1={x1}
                        x2={x2}
                        y1={medianY}
                        y2={medianY}
                        className="median" />
                    <line key="min"
                        x1={x1+2}
                        x2={x2-2}
                        y1={minY}
                        y2={minY}
                        className="tick" />
                </g>
            );
        };

        var yMaxX = padLeft - 3,
            yMaxY = padTop + 5;

        return (
            <div className="boxplot">
                <svg>
                    <text key="title" x={-70} y={10} textAnchor="middle" transform="rotate(270)" className="axis-label">Speed</text>
                    {this.props.points.map(renderBox)}
                    <line key="y-axis" x1={padLeft} x2={padLeft} y1={0} y2={maxHeight} className="axis" />
                    <path key="y-arrow" d="M30,0 32,7 28,7 Z" className="arrow" />
                    <text key="y-max" x={yMaxX} y={yMaxY} textAnchor="end"className="axis-label">{max.toFixed(1)}</text>
                    <text key="y-zero" x={27} y={100} textAnchor="end" className="axis-label">0</text>
                    <line key="x-axis" x1={padLeft} x2={950} y1={maxHeight} y2={maxHeight} className="axis" />
                    <path key="x-arrow" d="M950,140 943,142 943,138 Z" className="arrow" />
                    <text key="x-label-now" x={940} y={chartHeight - padBottom} textAnchor="end" className="axis-label">now</text>
                </svg>
            </div>
        );
    }
});

var LineChart = React.createClass({
    render: function() {
        var points = this.props.points,
            chartWidth = 950,
            chartHeight = 120,
            padLeft = 30,
            padTop = 10,
            padBottom = 20,
            valueInterval = 15;
            maxHeight = chartHeight - padTop - padBottom;

        var max = 0;
        points.map(function(point) {
            if (max === undefined || point.processed > max) {
                max = point.processed;
            }
        });

        var makePath = function(points, key) {
            if (max === 0) {
                return;
            }

            var npoints = points.map(function(point, i) {
                var val = point[key];
                var x = i * valueInterval + padLeft;
                var y = maxHeight - Math.round(val/max * maxHeight) + padTop;

                return [x, y];
            });

            var maxPointsRatio = points.map(function(point, i) {
                var val = point[key];
                return val === max ? 1 : 0;
            }).reduce(function(sum, val) {
                return sum + val;
            }) / points.length;

            var path = npoints.map(function(point, i) {
                var x = point[0], y = point[1];

                if (i === 0) {
                    return "M"+x+","+y;
                } else {
                    return "L"+x+","+y;
                }
            }).join(" ");

            var dots = npoints.map(function(point, i) {
                var x = point[0], y = point[1];

                var r = 2; // Radius
                // Hide leftmost and zero points
                if (x === padLeft || y === chartHeight - padBottom) {
                    r = 0;
                }
                // Highlight max values if less then 25% of values are max
                if (y == padTop && maxPointsRatio <= .25) {
                    r = 4;
                }

                return <circle key={"dot-"+ i}
                    cx={x}
                    cy={y}
                    r={r}
                    className={"dot "+ key} />
            });

            return (
                <g key={key}>
                    <path
                        d={path}
                        className={"line "+ key} />
                    {dots}
                </g>
            );
        };

        // TODO: Define magic numbers from below here
        var yMaxX = padLeft - 3,
            yMaxY = padTop + 5;

        var xlabels = Array.apply(null, Array(10)).map(function(_, i){
            return <text key={"x-label-"+ i}
                x={padLeft + (15 * 6 * i)}
                y={110}
                textAnchor="middle"
                className="axis-label">
                {"-"+ (10-i) + "m"}
            </text>
        });

        return (
            <div className="linechart">
                <svg>
                    <text key="title" x={-50} y={10} textAnchor="middle" transform="rotate(270)" className="axis-label">Throughput</text>
                    {makePath(points, "processed")}
                    {makePath(points, "errors")}
                    <line key="y-axis" x1={padLeft} x2={padLeft} y1={0} y2={100} className="axis" />
                    <path key="y-arrow" d="M30,0 32,7 28,7 Z" className="arrow" />
                    <text key="y-max" x={yMaxX} y={yMaxY} textAnchor="end"className="axis-label">{max}</text>
                    <text key="y-zero" x={27} y={100} textAnchor="end" className="axis-label">0</text>
                    <line key="x-axis" x1={30} x2={950} y1={100} y2={100} className="axis" />
                    <path key="x-arrow" d="M950,100 943,102 943,98 Z" className="arrow" />
                    <text key="x-label-now" x={940} y={110} textAnchor="end" className="axis-label">now</text>
                    {xlabels}
                </svg>
            </div>
        );
    }
});

ReactDOM.render(<Dashboard />, document.getElementById("app"));
