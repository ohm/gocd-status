/* eslint-env browser */

/* Copyright (c) 2016 Sebastian Ohm.
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

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
    var canvas = document.getElementById("pipelines"),
        context = canvas.getContext("2d"),
        cols = Math.ceil(pipelines.length / 6),
        rows = Math.ceil(pipelines.length / cols);

    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    var w = canvas.width / cols,
        h = canvas.height / rows,
        k = 0;

    for (var i = 0; i < cols; i++) {
        for (var j = 0; j < rows; j++) {
            k = (i * rows) + j;

            if (k >= pipelines.length) {
                context.fillStyle = "black";
                context.fillRect(i * w, j * h, (i * w) + w, (j * h) + h);
                continue;
            };

            context.fillStyle = "darkgray";

            var history = histories[pipelines[k]];
            if (history !== undefined) {
                if (history.length > 0) {
                    switch (history[0].Result) {
                        case "Unknown":
                            context.fillStyle = "orange";
                            break;
                        case "Failed":
                            context.fillStyle = "red";
                            break;
                        case "Passed":
                            context.fillStyle = "green";
                            break;
                    }
                }
            }

            context.fillRect(i * w, j * h, (i * w) + w, (j * h) + h);
            context.strokeStyle = "black";
            context.strokeRect(i * w, j * h, (i * w) + w, (j * h) + h);

            context.font = "24px sans-serif";
            context.fillStyle = "white";
            context.fillText(pipelines[k], (i * w) + 8, (j * h) + 24);
        }
    }
};

function updatePipelineGroup(name, done) {
    getJSON("/api/pipeline_groups.json", function(groups) {
        var group = groups.find(function(e) {
                return e.Name == name;
            }),
            histories = {};

        if (group === undefined) {
            console.log("Group " + name + " doesn't exist");
            done();
        };

        if (group.Pipelines.length == 0) {
            console.log("Group " + name + " doesn't have any pipelines");
            done();
        }

        var progress = function() {
            var queued = group.Pipelines.length;

            return function() {
                queued--;

                if (queued <= 0) {
                    drawPipelineGroup(group.Pipelines, histories);
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
        setTimeout(function() {
            update(delay);
        }, delay);
    });
};

var updateFn = updatePipelineGroups,
    group = decodeURI(window.location.pathname).split("/").pop();

if (group.length > 0) {
    updateFn = function(done) {
        updatePipelineGroup(group, done);
    };
}

update(5000);