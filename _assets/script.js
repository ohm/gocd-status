/* eslint-env browser */

/* Copyright (c) 2016 Sebastian Ohm.
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

var Layout = function(canvas) {
    this.canvas = canvas;
    this.context = canvas.getContext("2d");
};

Layout.prototype.Init = function() {
    this.canvas.width = window.innerWidth;
    this.canvas.height = window.innerHeight;
};

Layout.prototype.Draw = function(pipelines) {
    var cols = getQueryParam("cols", Math.ceil(pipelines.length / 6)),
        rows = Math.ceil(pipelines.length / cols),
        w = this.canvas.width / cols,
        h = this.canvas.height / rows,
        k = 0,
        x = 0,
        y = 0;

    for (var i = 0; i < cols; i++) {
        for (var j = 0; j < rows; j++) {
            k = (i * rows) + j;
            x = i * w;
            y = j * h;

            this.Prepare(x, y, w, h);

            if (k < pipelines.length) {
                this.Pipeline(x, y, w, h, pipelines[k]);
            }
        }
    }
};

Layout.prototype.Prepare = function(x, y, w, h) {
    this.context.fillStyle = "black";
    this.context.fillRect(x, y, x + w, y + h);
};

Layout.prototype.Pipeline = function(x, y, w, h, pipeline) {
    var colors = pipeline.Histories.map(function(h) {
        switch(h.Result) {
            case "Failed":
                return "red";
            case "Passed":
                return "green";
            case "Unknown":
                // TODO: Building?
                return "orange";
            default:
                return "darkgray";
        }
    });

    this.context.fillStyle = colors.shift() || "darkgray";
    this.context.fillRect(x, y, x + w, y + h);

    this.context.strokeStyle = "black";
    this.context.strokeRect(x, y, x + w, y + h);

    this.context.font = "24px sans-serif";
    this.context.fillStyle = "white";
    this.context.fillText(pipeline.Name, x + 8, y + 24);
};

function getQueryParam(key, value) {
    var val = window.location.search
        .substring(1)
        .split('&')
        .map(function(x) { return x.split("=") })
        .filter(function(x) { return x[0] == key })
        .map(function(x) { return parseInt(x[1]) })
        .pop();

    return isNaN(val) ? value : val;
};

function getJSON(url, success) {
    var request = new XMLHttpRequest();

    request.open("GET", url, true);
    request.onload = function() {
        if (request.status >= 200 && request.status < 400) {
            success(JSON.parse(request.responseText));
        }
    };
    request.send();
};

function drawPipelineGroups(groups) {
    var html = groups.map(function(g) {
        return "<li><a href=\"/" + g.Name + "\">" + g.Name + " (" + g.Pipelines.length + ")</a></li>";
    });

    document
        .getElementById("pipeline-groups")
        .innerHTML = html.join("");
};

function updatePipelineGroups(done) {
    getJSON("/api/pipeline_groups.json", function(groups) {
        drawPipelineGroups(groups);
        done();
    });
};

function drawPipelineGroup(pipelines, histories) {
    var l = new Layout(document.getElementById("pipelines"));

    l.Init();
    l.Draw(pipelines, histories);
};

function updatePipelineGroup(name, done) {
    getJSON("/api/pipeline_groups.json", function(groups) {
        var group = groups.find(function(e) { return e.Name == name }),
            histories = {};

        if (group === undefined) {
            console.log("Group " + name + " doesn't exist");
            done();
        }

        if (group.Pipelines.length == 0) {
            console.log("Group " + name + " doesn't have any pipelines");
            done();
        }

        var progress = function() {
            var queued = group.Pipelines.length;

            return function() {
                queued--;

                if (queued <= 0) {
                    drawPipelineGroup(group.Pipelines.map(function(name) {
                        return {
                          "Name": name,
                          "Histories": histories[name] || []
                        };
                    }));
                    done();
                };
            };
        }();

        group.Pipelines.forEach(function(name) {
            getJSON("/api/pipeline_history.json?pipeline=" + name, function(history) {
                histories[name] = history;
                progress();
            });
        });
    });
};

function update(delay) {
    updateFn(function() {
        setTimeout(function() { update(delay) }, delay);
    });
};

var updateFn = updatePipelineGroups,
    group = decodeURI(window.location.pathname).split("/").pop();

if (group.length > 0) {
    updateFn = function(done) { updatePipelineGroup(group, done) };
}

update(getQueryParam("refresh", 5000));
