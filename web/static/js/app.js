document.addEventListener("DOMContentLoaded", () => {
    const revealItems = document.querySelectorAll(".reveal");
    const counters = document.querySelectorAll("[data-counter]");

    const revealObserver = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
            if (entry.isIntersecting) {
                entry.target.classList.add("is-visible");
                revealObserver.unobserve(entry.target);
            }
        });
    }, { threshold: 0.16 });

    revealItems.forEach((item) => revealObserver.observe(item));

    const counterObserver = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
            if (!entry.isIntersecting) {
                return;
            }

            const node = entry.target;
            const target = Number.parseFloat(node.dataset.counter || "0");
            const isDecimal = !Number.isInteger(target);
            const duration = 1200;
            const start = performance.now();
            const finalText = isDecimal ? target.toFixed(2) : target.toString();

            function tick(now) {
                const progress = Math.min((now - start) / duration, 1);
                const eased = 1 - Math.pow(1 - progress, 3);
                const value = target * eased;

                node.textContent = isDecimal ? value.toFixed(2) : Math.round(value).toString();

                if (progress < 1) {
                    requestAnimationFrame(tick);
                } else {
                    node.textContent = finalText;
                }
            }

            node.textContent = "0";
            requestAnimationFrame(tick);
            counterObserver.unobserve(node);
        });
    }, { threshold: 0.45 });

    counters.forEach((counter) => counterObserver.observe(counter));
});
