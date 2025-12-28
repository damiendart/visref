// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

window.customElements.define(
  'message-banner',
  class extends HTMLElement {
    connectedCallback() {
      this.addEventListener(
        "click",
        (event) => {
          if (event.target.hasAttribute("href") === false) {
            return;
          }

          const href = event.target.getAttribute("href");

          if (href?.startsWith("#")) {
            const el = document.getElementById(href.substring(1));

            if (el === null) {
              return;
            }

            event.preventDefault();

            if (el.closest(".form-item")) {
              el.closest(".form-item").scrollIntoView();
            } else {
              el.scrollIntoView();
            }
          }
        },
      );
    }
  },
);
