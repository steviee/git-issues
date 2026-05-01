import { useState } from 'react';
import { Copy, Check } from 'lucide-react';

export default function InstallCommand() {
  const [copied, setCopied] = useState(false);

  const command = 'go install github.com/steviee/git-issues@latest';

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback: silently ignore
    }
  };

  return (
    <div className="flex items-center gap-2 bg-[#0c0c0c] border border-[#1f1f1f] rounded-lg px-4 py-3 max-w-fit">
      <span className="font-mono text-sm text-[#e5e5e5] whitespace-nowrap">
        {command}
      </span>
      <button
        onClick={handleCopy}
        className="ml-2 p-1.5 rounded-md text-[#525252] hover:text-[#a3a3a3] hover:bg-[#1f1f1f] transition-colors flex items-center gap-1.5"
        aria-label="Copy install command"
      >
        {copied ? (
          <>
            <Check size={14} className="text-[#22c55e]" />
            <span className="text-xs text-[#22c55e] font-medium">Copied!</span>
          </>
        ) : (
          <Copy size={14} />
        )}
      </button>
    </div>
  );
}
