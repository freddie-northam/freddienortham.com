/* Four motion modules, each self-registering off a data attribute. Deleting
   one means deleting one function and one attribute; nothing else breaks.

   Under prefers-reduced-motion every module renders its static end state and
   registers no listeners at all. */

(() => {
  "use strict";

  const reduced = matchMedia("(prefers-reduced-motion: reduce)").matches;
  const coarse = matchMedia("(pointer: coarse)").matches;

  /* 1. Echo stack -------------------------------------------------------- */
  /* The shadow stack rises as one. Per-ghost stagger is not possible because
     browsers interpolate a text-shadow list layer by layer, which is the
     trade we accepted to avoid duplicating the title four times in the DOM. */
  function echo() {
    const nodes = document.querySelectorAll("[data-echo]");
    if (!nodes.length) return;
    if (reduced) {
      nodes.forEach((n) => n.classList.add("is-in"));
      return;
    }
    const io = new IntersectionObserver(
      (entries) => {
        entries.forEach((e) => {
          if (!e.isIntersecting) return;
          e.target.classList.add("is-in");
          io.unobserve(e.target);
        });
      },
      { rootMargin: "0px 0px -10% 0px" }
    );
    nodes.forEach((n) => io.observe(n));
  }

  /* 2. Weight wave ------------------------------------------------------- */
  /* Geist varies on weight, not width, so letter advance widths never change
     and the word cannot reflow while animating.

     Mobile has no cursor. Without the coarse-pointer branch the site's
     signature moment would simply not exist for most visitors, so there the
     wave centre travels the word on a loop instead. */
  function wave() {
    const el = document.querySelector("[data-wave]");
    if (!el) return;
    const spans = [...el.querySelectorAll(".ltr")];
    if (!spans.length) return;

    const MIN = 200;
    const MAX = 900;
    const SIGMA = 90;
    const REST = 800;

    if (reduced) {
      spans.forEach((s) => (s.style.fontVariationSettings = `"wght" ${REST}`));
      return;
    }

    const current = spans.map(() => REST);
    let target = null; // x in page coords; null means rest
    let running = false;
    let visible = false;
    let t0 = 0;

    const centres = () =>
      spans.map((s) => {
        const r = s.getBoundingClientRect();
        return r.left + r.width / 2;
      });

    function frame(now) {
      if (!running) return;
      const cs = centres();

      if (coarse) {
        // Autonomous sweep: 3.2s across the word, easing at both ends.
        if (!t0) t0 = now;
        const phase = ((now - t0) % 3200) / 3200;
        const eased = 0.5 - 0.5 * Math.cos(phase * 2 * Math.PI);
        const first = cs[0];
        const last = cs[cs.length - 1];
        target = first + (last - first) * eased;
      }

      let moved = false;
      spans.forEach((s, i) => {
        let want = REST;
        if (target !== null) {
          const d = cs[i] - target;
          const g = Math.exp(-(d * d) / (2 * SIGMA * SIGMA));
          want = MIN + (MAX - MIN) * g;
        }
        const next = current[i] + (want - current[i]) * 0.15;
        if (Math.abs(next - current[i]) > 0.4) moved = true;
        current[i] = next;
        s.style.fontVariationSettings = `"wght" ${Math.round(next)}`;
      });

      if (!moved && target === null) {
        running = false;
        return;
      }
      requestAnimationFrame(frame);
    }

    function start() {
      if (running || !visible || document.hidden) return;
      running = true;
      requestAnimationFrame(frame);
    }

    const io = new IntersectionObserver((entries) => {
      visible = entries[0].isIntersecting;
      if (visible) start();
      else running = false;
    });
    io.observe(el);

    document.addEventListener("visibilitychange", () => {
      if (document.hidden) running = false;
      else start();
    });

    if (coarse) {
      el.addEventListener(
        "touchstart",
        (e) => {
          t0 = 0;
          target = e.touches[0].clientX;
          start();
        },
        { passive: true }
      );
    } else {
      el.addEventListener("pointermove", (e) => {
        target = e.clientX;
        start();
      });
      el.addEventListener("pointerleave", () => {
        target = null;
        start();
      });
    }
  }

  /* 3. Hover grid -------------------------------------------------------- */
  /* Pure CSS on :hover. The only job here is to switch it off where hover
     does not exist, since a tap would otherwise leave a row stuck lit. */
  function grid() {
    if (!coarse && !reduced) return;
    document.querySelectorAll("[data-grid]").forEach((n) => {
      n.querySelectorAll("a").forEach((a) => (a.style.setProperty("--grid-off", "1")));
      n.classList.add("no-grid");
    });
    const style = document.createElement("style");
    style.textContent = ".no-grid .row > a::before{display:none}";
    document.head.appendChild(style);
  }

  /* 4. Days on earth ----------------------------------------------------- */
  /* Below the fold on every page, so it must cost nothing when nobody is
     looking: gated on both intersection and tab visibility. A naive interval
     would write to the DOM ten times a second, forever, unobserved. */
  function days() {
    const el = document.querySelector("[data-days]");
    if (!el) return;
    const zero = Number(el.dataset.days) * 1000;
    if (!zero) return;

    const DAY = 86400000;
    const render = (frac) => {
      const d = (Date.now() - zero) / DAY;
      el.textContent = frac ? d.toFixed(7) + " days" : Math.floor(d).toLocaleString() + " days";
    };

    if (reduced) {
      render(false);
      return;
    }

    let visible = false;
    let running = false;

    function tick() {
      if (!running) return;
      render(true);
      requestAnimationFrame(tick);
    }
    function update() {
      const go = visible && !document.hidden;
      if (go === running) return;
      running = go;
      if (go) requestAnimationFrame(tick);
    }

    render(true);
    new IntersectionObserver((entries) => {
      visible = entries[0].isIntersecting;
      update();
    }).observe(el);
    document.addEventListener("visibilitychange", update);
  }

  /* Article furniture ---------------------------------------------------- */
  function furniture() {
    document.querySelectorAll("[data-anchors] h2[id], [data-anchors] h3[id]").forEach((h) => {
      const a = document.createElement("a");
      a.className = "anchor";
      a.href = "#" + h.id;
      a.setAttribute("aria-label", `Link to ${h.textContent}`);
      a.innerHTML =
        '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1"/></svg>';
      h.appendChild(a);
    });

    document.querySelectorAll(".prose pre").forEach((pre) => {
      const wrap = document.createElement("div");
      wrap.className = "pre-wrap";
      pre.parentNode.insertBefore(wrap, pre);
      wrap.appendChild(pre);

      const btn = document.createElement("button");
      btn.className = "copy-btn";
      btn.type = "button";
      btn.setAttribute("aria-label", "Copy code");
      btn.innerHTML =
        '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg>';
      btn.addEventListener("click", async () => {
        await navigator.clipboard.writeText(pre.innerText);
        btn.setAttribute("aria-label", "Copied");
        setTimeout(() => btn.setAttribute("aria-label", "Copy code"), 1200);
      });
      wrap.appendChild(btn);
    });
  }

  echo();
  wave();
  grid();
  days();
  furniture();
})();
