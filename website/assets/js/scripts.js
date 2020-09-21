var body = document.querySelector('body')
var menuTrigger = document.querySelector('#toggle-main-menu-mobile');
var menuContainer = document.querySelector('#main-menu-mobile');

menuTrigger.onclick = function() {
    menuContainer.classList.toggle('open');
    menuTrigger.classList.toggle('is-active')
    body.classList.toggle('lock-scroll')
}

// $('body').scrollspy({
//     target: '.bs-docs-sidebar',
//     offset: 40
// });


var content = document.querySelector('.content.anchor-link-enabled')
if (content) {
    addHeaderAnchors(content);
}

function addHeaderAnchors(content) {
    var headers = content.querySelectorAll('h1, h2, h3, h4');
    // SVG data from https://iconmonstr.com/link-1-svg/
    var linkSvg = ' <svg xmlns="http://www.w3.org/2000/svg" width="16px" height="16px" viewBox="0 0 24 24"><path d="M0 0h24v24H0z" fill="none"></path><path d="M6.188 8.719c.439-.439.926-.801 1.444-1.087 2.887-1.591 6.589-.745 8.445 2.069l-2.246 2.245c-.644-1.469-2.243-2.305-3.834-1.949-.599.134-1.168.433-1.633.898l-4.304 4.306c-1.307 1.307-1.307 3.433 0 4.74 1.307 1.307 3.433 1.307 4.74 0l1.327-1.327c1.207.479 2.501.67 3.779.575l-2.929 2.929c-2.511 2.511-6.582 2.511-9.093 0s-2.511-6.582 0-9.093l4.304-4.306zm6.836-6.836l-2.929 2.929c1.277-.096 2.572.096 3.779.574l1.326-1.326c1.307-1.307 3.433-1.307 4.74 0 1.307 1.307 1.307 3.433 0 4.74l-4.305 4.305c-1.311 1.311-3.44 1.3-4.74 0-.303-.303-.564-.68-.727-1.051l-2.246 2.245c.236.358.481.667.796.982.812.812 1.846 1.417 3.036 1.704 1.542.371 3.194.166 4.613-.617.518-.286 1.005-.648 1.444-1.087l4.304-4.305c2.512-2.511 2.512-6.582.001-9.093-2.511-2.51-6.581-2.51-9.092 0z"/></svg>';
    var anchorForId = function (id) {
        var anchor = document.createElement('a');
        anchor.classList.add('header-anchor');
        anchor.href = "#" + id;
        anchor.innerHTML = linkSvg;
        return anchor;
    };

    for (var h = 0; h < headers.length; h++) {
        var header = headers[h];

        if (typeof header.id !== "undefined" && header.id !== "") {
            header.appendChild(anchorForId(header.id));
        }
    }
}


// SIDEBAR
// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.




function getById(id) {
    return document.getElementById(id);
}

function listen(o, e, f) {
    if (o) {
        o.addEventListener(e, f);
    }
}

function toggleAttribute(el, name) {
    if (el.getAttribute(name) === "true") {
        el.setAttribute(name, "false");
    } else {
        el.setAttribute(name, "true");
    }
}

const click = "click";
const mouseenter = "mouseenter";
const mouseleave = "mouseleave";
const active = "active";
const keyup = "keyup";
const keydown = "keydown";
const button = "button";
const ariaLabel = "aria-label";
const ariaExpanded = "aria-expanded";
const ariaSelected = "aria-selected";
const ariaControls = "aria-controls";
const tabIndex = "tabindex";

// Attach the event handlers to support the sidebar
function handleSidebar() {
    const sidebar = getById("sidebar");
    if (!sidebar) {
        return;
    }

    // toggle subtree in sidebar
    sidebar.querySelectorAll(".body").forEach(body => {
        body.querySelectorAll(button).forEach(o => {
            listen(o, click, e => {
                const button = e.currentTarget;
                button.classList.toggle("show");
                const next = button.nextElementSibling;
                if (!next) {
                    return;
                }

                const ul = next.nextElementSibling;
                if (!ul) {
                    return;
                }

                toggleAttribute(ul, ariaExpanded);

                let el = ul;
                do {
                    el = el.parentElement;
                } while (!el.classList.contains("body"));

                // adjust the body's max height to the total size of the body's content
                el.style.maxHeight = el.scrollHeight + "px";
            });
        });

        // window.observeResize(body, el => {
        //     if ((el.style.maxHeight !== null) && (el.style.maxHeight !== "")) {
        //         el.style.maxHeight = el.scrollHeight + "px";
        //     }
        // });
    });

    const headers = [];
    sidebar.querySelectorAll(".header").forEach(header => {
        headers.push(header);
    });

    function toggleHeader(header) {
        const body = header.nextElementSibling;
        if (!body) {
            return;
        }

        body.classList.toggle("show");
        toggleAttribute(header, ariaExpanded);

        if (body.classList.contains("show")) {
            // set this as the limit for expansion
            body.style.maxHeight = body.scrollHeight + "px";
        } else {
            // if was expanded, reset this
            body.style.maxHeight = "";
        }
    }

    // expand/collapse cards
    sidebar.querySelectorAll(".header").forEach(header => {
        if (header.classList.contains("dynamic")) {
            listen(header, click, () => {
                toggleHeader(header);
            });
        }
    });

    // force expand the default cards
    sidebar.querySelectorAll(".body").forEach(body => {
        if (body.classList.contains("default")) {
            body.style.maxHeight = body.scrollHeight + "px";
            body.classList.toggle("default");
            body.classList.toggle("show");
            const header = body.previousElementSibling;
            if (header) {
                toggleAttribute(header, ariaExpanded);
            }
        }
    });

    // toggle sidebar on/off
    listen(getById("sidebar-toggler"), click, e => {
        const sc = getById("sidebar-container");
        if (sc) {
            sc.classList.toggle(active);
            const icon = (e.currentTarget).querySelector("svg.icon");
            if (icon) {
                icon.classList.toggle("flipped");
            }
        }
    });
}

handleSidebar();