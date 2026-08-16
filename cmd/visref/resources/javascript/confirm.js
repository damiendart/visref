// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

document.addEventListener(
  "click",
  (event) => {
    if (
      event.target.hasAttribute("data-confirm")
      && !window.confirm(event.target.dataset.confirm)
    ) {
      event.preventDefault();
    }
  },
);
