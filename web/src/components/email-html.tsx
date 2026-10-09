import { useEffect, useMemo, useRef, useState } from "react";
import { ImageOff } from "lucide-react";

const BASE_STYLE = `
  html, body { margin: 0; padding: 0; }
  body { font-family: system-ui, -apple-system, "Segoe UI", sans-serif; font-size: 14px; line-height: 1.5; color: #111; word-wrap: break-word; overflow-wrap: anywhere; }
  img { max-width: 100%; height: auto; }
  pre { white-space: pre-wrap; }
`;

// Elements that can navigate, embed other documents or run code. Scripts are
// already blocked by the sandbox and CSP; this strips the rest.
const STRIP = "script, iframe, frame, frameset, object, embed, applet, form, base, link, meta, noscript";

// Parses the email with DOMParser (inert: nothing loads or runs) and returns
// the cleaned <style> blocks and body, plus whether it references remote content.
function sanitize(html: string) {
  const doc = new DOMParser().parseFromString(html, "text/html");
  doc.querySelectorAll(STRIP).forEach((el) => el.remove());
  doc.querySelectorAll("*").forEach((el) => {
    for (const attr of Array.from(el.attributes)) {
      if (attr.name.startsWith("on")) el.removeAttribute(attr.name);
    }
  });
  doc.querySelectorAll("a").forEach((a) => {
    a.setAttribute("target", "_blank");
    a.setAttribute("rel", "noopener noreferrer");
  });
  const styles = Array.from(doc.querySelectorAll("style"))
    .map((s) => s.outerHTML)
    .join("");
  doc.querySelectorAll("style").forEach((s) => s.remove());
  const body = doc.body?.innerHTML ?? "";
  const hasRemote = /\b(?:src|background|srcset)\s*=\s*["']?\s*(?:https?:)?\/\/|url\(\s*["']?\s*(?:https?:)?\/\//i.test(html);
  return { styles, body, hasRemote };
}

// Renders untrusted email HTML in an isolated iframe. Scripts are blocked
// twice (no allow-scripts, CSP script-src 'none'), forms and navigation are
// stripped, and remote images/CSS are blocked until the reader opts in so
// senders can't use tracking pixels to learn when and where mail was opened.
export default function EmailHtml({ html }: { html: string }) {
  const ref = useRef<HTMLIFrameElement>(null);
  const [height, setHeight] = useState(0);
  const [showRemote, setShowRemote] = useState(false);

  const { styles, body, hasRemote } = useMemo(() => sanitize(html), [html]);
  const remoteSrc = showRemote ? " https:" : "";

  const srcDoc = `<!doctype html><html><head><meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'none'; object-src 'none'; base-uri 'none'; form-action 'none'; style-src 'unsafe-inline'; img-src data: cid:${remoteSrc}; font-src data:">
<meta name="referrer" content="no-referrer">
<style>${BASE_STYLE}</style>${styles}</head><body>${body}</body></html>`;

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
    <div className="space-y-2">
      {hasRemote && !showRemote && (
        <div className="flex flex-wrap items-center gap-2 rounded-lg border bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
          <ImageOff className="h-3.5 w-3.5" />
          Remote images are hidden to protect your privacy.
          <button type="button" className="font-medium text-foreground underline" onClick={() => setShowRemote(true)}>
            Show images
          </button>
        </div>
      )}
      <iframe
        ref={ref}
        title="Email content"
        srcDoc={srcDoc}
        // allow-same-origin only lets us measure height; scripts stay disabled.
        sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
        referrerPolicy="no-referrer"
        className="w-full max-w-full block border-0 bg-white rounded"
        style={{ height: height || 200 }}
      />
    </div>
  );
}
