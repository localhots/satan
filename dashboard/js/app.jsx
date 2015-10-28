var Dashboard = React.createClass({
    getInitialState: function() {
        return {};
    },

    componentDidMount: function() {
        this.reload();
    },

    reload: function() {
        getURL("http://127.0.0.1:6464/stats.json", {}, function(resp) {
            this.setState(resp);
            setTimeout(this.reload, 5000);
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
    decode: function(point) {
        return {
            timestamp: point[0],
            processed: point[1],
            errors:    point[2],
            min:       point[3],
            mean:      point[4],
            p95:       point[5],
            max:       point[6],
            stddev:    point[7],
        }
    },

    render: function() {
        var last = this.decode(this.props.points[this.props.points.length - 1]);
        return (
            <div className="daemon">
                <h1>{this.props.name}</h1>
                <dl className="stats">
                    <dt>processed:</dt><dd>{last.processed}</dd>
                    <dt className="narrow">errors:</dt><dd>{last.errors}  </dd>
                    <dt>min:</dt><dd>{formatDuration(last.min)}</dd>
                    <dt className="narrow">max:</dt><dd>{formatDuration(last.max)}</dd>
                    <dt>mean:</dt><dd>{formatDuration(last.mean)}</dd>
                    <dt className="narrow">95%:</dt><dd>{formatDuration(last.p95)}</dd>
                </dl>
            </div>
        );
    }
});

ReactDOM.render(<Dashboard />, document.getElementById("app"));
