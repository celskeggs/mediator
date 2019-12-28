"use strict";

function StatPanel(tabdiv, bodydiv, imageLoader) {
    this.tabdiv = tabdiv;
    this.bodydiv = bodydiv;
    this.statPanels = {};
    this.verbs = [];
    this.imageLoader = imageLoader;
    this.onverb = null;
    this.oncontextmenu = null;
    this.selectedStatPanel = null;
}

function updateChildren(target, entries, newElement, updateElement) {
    while (target.children.length > entries.length) {
        target.children[target.children.length - 1].remove();
    }
    while (target.children.length < entries.length) {
        target.appendChild(newElement());
    }
    for (let i = 0; i < entries.length; i++) {
        updateElement(target.children[i], entries[i]);
    }
}

function removeChildren(target) {
    while (target.children.length > 0) {
        target.children[0].remove();
    }
}

// returns true if the class was changed
function updateCSSClass(target, include, cssClass, notClass) {
    if (include) {
        if (notClass) {
            target.classList.remove(notClass);
        }
        if (!target.classList.contains(cssClass)) {
            target.classList.add(cssClass);
            return true;
        }
    } else {
        if (notClass) {
            target.classList.add(notClass);
        }
        if (target.classList.contains(cssClass)) {
            target.classList.remove(cssClass);
            return true;
        }
    }
    return false;
}

StatPanel.prototype.renderTabs = function () {
    const tabs = this.listPanels();
    const panel = this;
    updateChildren(this.tabdiv, tabs, function () {
        const button = document.createElement("button");
        button.addEventListener("click", function () {
            panel.selectedStatPanel = this.textContent;
            panel.rerender();
        });
        return button;
    }, function (elem, tab) {
        elem.textContent = tab;
        updateCSSClass(elem, tab === panel.selectedStatPanel, "selected");
    });
};

StatPanel.prototype.renderBody = function (entries, areVerbs) {
    const panel = this;
    if (updateCSSClass(this.bodydiv, areVerbs, "verbpanel", "statpanel")) {
        // if we changed modes, we'll need to wipe and re-create the entries
        removeChildren(this.bodydiv);
    }
    if (areVerbs) {
        updateChildren(this.bodydiv, entries, function () {
            const child = document.createElement("div");
            child.addEventListener("click", function () {
                if (panel.onverb !== null) {
                    panel.onverb(this.textContent);
                }
            });
            return child;
        }, function (elem, entry) {
            elem.textContent = entry;
        });
    } else {
        updateChildren(this.bodydiv, entries, function () {
            const child = document.createElement("div");
            const statlabel = document.createElement("span");
            statlabel.classList.add("statlabel");
            child.appendChild(statlabel);
            const object = document.createElement("div");
            object.appendChild(document.createElement("div"));
            object.appendChild(document.createElement("span"));
            object.entry = null;
            object.addEventListener("contextmenu", function (ev) {
                if (panel.oncontextmenu !== null && this.entry !== null) {
                    panel.oncontextmenu(ev, this.entry);
                }
            });
            child.appendChild(object);
            const statsuffix = document.createElement("span");
            statsuffix.classList.add("statsuffix");
            child.appendChild(statsuffix);
            return child;
        }, function (elem, entry) {
            const labelSpan = elem.children[0];
            const statObject = elem.children[1];
            const iconDiv = statObject.children[0];
            const nameSpan = statObject.children[1];
            const suffixSpan = elem.children[2];

            updateCSSClass(statObject, entry.verbs && entry.verbs.length > 0, "statobject");
            statObject.entry = entry;

            labelSpan.textContent = entry.label;
            nameSpan.textContent = entry.name;
            suffixSpan.textContent = entry.suffix;

            panel.imageLoader.updateHTMLIcon(entry, iconDiv);
        });
    }
};

StatPanel.prototype.listPanels = function () {
    const allPanels = [];
    for (const panel in this.statPanels) {
        if (this.statPanels.hasOwnProperty(panel)) {
            allPanels.push(panel);
        }
    }
    allPanels.sort();
    if (this.verbs.length > 0) {
        allPanels.push("Commands");
    }
    return allPanels;
};

StatPanel.prototype.rerender = function () {
    const allPanels = this.listPanels();
    if (allPanels.indexOf(this.selectedStatPanel) === -1) {
        this.selectedStatPanel = allPanels.length > 0 ? allPanels[0] : null;
    }
    this.renderTabs();
    if (this.selectedStatPanel === "Commands") {
        this.renderBody(this.verbs, true);
    } else if (this.selectedStatPanel !== null) {
        const paneldata = this.statPanels[this.selectedStatPanel];
        this.renderBody(paneldata.entries || [], false);
    } else {
        this.renderBody([], false);
    }
};

StatPanel.prototype.update = function (verbs, panels) {
    this.verbs = verbs;
    this.statPanels = panels;
    this.rerender();
};
