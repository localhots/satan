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
            <div className="daemons">{this.renderDaemons()}</div>
        );
    }
});

var Daemon = React.createClass({
    render: function() {
        var last = this.props.points[this.props.points.length - 1];
        return (
            <div className="daemon">
                <div className="left-block">
                    <h1>{this.props.name}</h1>
                    <dl>
                        <dt>Processed:</dt><dd>{last.processed}</dd>
                        <dt>Errors:</dt><dd>{last.errors}</dd>
                        <dt>Median:</dt><dd>{formatDuration(last.median)}</dd>
                    </dl>
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
                    <line key="min-bar"
                        x1={x1+width/2}
                        x2={x1+width/2}
                        y1={minY}
                        y2={p25Y}
                        strokeDasharray="3,1"
                        strokeWidth={1}
                        style={{stroke: "#ccc"}} />
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

ReactDOM.render(<Dashboard />, document.getElementById("app"));
