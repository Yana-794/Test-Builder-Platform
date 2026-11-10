import { useEffect, useRef } from "react";


export function useMouseParallax() {
  const ref = useRef<HTMLElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    const reduced = matchMedia("(prefers-reduced-motion: reduce)").matches;
    const finePointer = matchMedia(
      "(hover: hover) and (pointer: fine)",
    ).matches;
    if (reduced || !finePointer) return;

    const target = { x: 0, y: 0 };
    const current = { x: 0, y: 0 };

    let raf = 0;

    const loop = () => {
      current.x += (target.x - current.x) * 0.08;
      current.y += (target.y - current.y) * 0.08;

      el.style.setProperty("--mx", current.x.toFixed(3));
      el.style.setProperty("--my", current.y.toFixed(3));

      const dx = Math.abs(target.x - current.x);
      const dy = Math.abs(target.y - current.y);

      if (dx > 0.001 || dy > 0.001) {
        raf = requestAnimationFrame(loop);
      } else {
        raf = 0;
      }
    };

    const onMove = (e: PointerEvent) => {
      target.x = (e.clientX / window.innerWidth) * 2 - 1;
      target.y = (e.clientY / window.innerHeight) * 2 - 1;

      if (!raf) raf = requestAnimationFrame(loop);
    };

    window.addEventListener("pointermove", onMove, { passive: true });

    return () => {
      window.removeEventListener("pointermove", onMove);
      if (raf) cancelAnimationFrame(raf);
    };
  }, []);

  return ref;
}
