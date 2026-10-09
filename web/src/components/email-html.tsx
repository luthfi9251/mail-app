import { useEffect, useRef, useState } from "react";

const BASE_STYLE = `
  html, body { margin: 0; padding: 0; }
  body { font-family: system-ui, -apple-system, "Segoe UI", sans-serif; font-size: 14px; line-height: 1.5; color: #111; word-wrap: break-word; overflow-wrap: anywhere; }
  img { max-width: 100%; height: auto; }
  pre { white-space: pre-wrap; }
`;

// Renders untrusted email HTML in an isolated iframe so its <style> rules and
// markup can't leak into (or break) the app. Scripts are blocked twice: no
// allow-scripts in the sandbox and a CSP that forbids them.
export default function EmailHtml({ html }: { html: string }) {
  const ref = useRef<HTMLIFrameElement>(null);
  const [height, setHeight] = useState(0);

  const srcDoc = `<!doctype html><html><head><meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="script-src 'none'; object-src 'none'; base-uri 'none'">
<base target="_blank"><style>${BASE_STYLE}</style></head><body>${html}</body></html>`;

  useEffect(() => {
    const frame = ref.current;
    if (!frame) return;
    let observer: ResizeObserver | undefined;

    function measure() {
      const doc = frame?.contentDocument;
      if (!doc?.documentElement) return;
      setHeight(Math.ceil(doc.documentElement.scrollHeight));
    }

    function onLoad() {
      measure();
      const doc = frame?.contentDocument;
      if (doc?.body && typeof ResizeObserver !== "undefined") {
        observer = new ResizeObserver(measure);
        observer.observe(doc.body);
      }
    }

    frame.addEventListener("load", onLoad);
    return () => {
      frame.removeEventListener("load", onLoad);
      observer?.disconnect();
    };
  }, [srcDoc]);

  return (
    <iframe
      ref={ref}
      title="Email content"
      srcDoc={srcDoc}
      // allow-same-origin only lets us measure height; scripts stay disabled.
      sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
      className="w-full max-w-full block border-0 bg-white rounded"
      style={{ height: height || 200 }}
    />
  );
}
