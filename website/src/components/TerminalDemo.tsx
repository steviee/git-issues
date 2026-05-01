import { useState, useEffect, useRef, useCallback } from 'react';

interface Line {
  type: 'command' | 'output' | 'table' | 'tree';
  content: string;
}

const DEMO_SEQUENCE: Line[] = [
  { type: 'command', content: '$ issues init' },
  { type: 'output', content: 'Initialized git-issues in .issues/' },
  { type: 'output', content: '' },
  { type: 'command', content: '$ issues new --title "Fix login bug" --priority critical --label bug' },
  { type: 'output', content: 'Created: .issues/0001-fix-login-bug.md (#1)' },
  { type: 'output', content: '' },
  { type: 'command', content: '$ issues new --title "Add dark mode" --label feature' },
  { type: 'output', content: 'Created: .issues/0002-add-dark-mode.md (#2)' },
  { type: 'output', content: '' },
  { type: 'command', content: '$ issues new --title "Session timeout" --priority high --label bug' },
  { type: 'output', content: 'Created: .issues/0003-session-timeout.md (#3)' },
  { type: 'output', content: '' },
  { type: 'command', content: '$ issues list' },
  { type: 'table', content: 'ID    PRI       STATUS       TITLE                        LABELS' },
  { type: 'table', content: '0003  high      open         Session timeout                [bug]' },
  { type: 'table', content: '0001  critical  open         Fix login bug                  [bug]' },
  { type: 'table', content: '0002  medium    open         Add dark mode                  [feature]' },
  { type: 'output', content: '' },
  { type: 'command', content: '$ issues graph --open-only' },
  { type: 'tree', content: '#3  Session timeout [open, high]' },
  { type: 'tree', content: '    (no relations)' },
  { type: 'output', content: '' },
  { type: 'tree', content: '#1  Fix login bug [open, critical]' },
  { type: 'tree', content: '    (no relations)' },
  { type: 'output', content: '' },
  { type: 'tree', content: '#2  Add dark mode [open, medium]' },
  { type: 'tree', content: '    (no relations)' },
];

const CHAR_SPEED = 20;
const LINE_PAUSE = 400;

export default function TerminalDemo() {
  const [displayLines, setDisplayLines] = useState<Line[]>([]);
  const [currentLine, setCurrentLine] = useState(0);
  const [currentChar, setCurrentChar] = useState(0);
  const [isTyping, setIsTyping] = useState(true);
  const [isPaused, setIsPaused] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const resetDemo = useCallback(() => {
    setDisplayLines([]);
    setCurrentLine(0);
    setCurrentChar(0);
    setIsTyping(true);
    setIsPaused(false);
  }, []);

  useEffect(() => {
    if (isPaused) return;

    if (currentLine >= DEMO_SEQUENCE.length) {
      // Demo finished, pause before restart
      setIsPaused(true);
      setTimeout(() => {
        resetDemo();
      }, 3000);
      return;
    }

    const line = DEMO_SEQUENCE[currentLine];
    const content = line.content;

    if (currentChar < content.length) {
      const timer = setTimeout(() => {
        setCurrentChar((prev) => prev + 1);
      }, CHAR_SPEED);
      return () => clearTimeout(timer);
    } else {
      // Line complete
      const timer = setTimeout(() => {
        setDisplayLines((prev) => [...prev, line]);
        setCurrentLine((prev) => prev + 1);
        setCurrentChar(0);
      }, LINE_PAUSE);
      return () => clearTimeout(timer);
    }
  }, [currentLine, currentChar, isPaused, resetDemo]);

  // Scroll to bottom
  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [displayLines, currentChar]);

  const renderContent = (line: Line, isActive: boolean, charIndex: number) => {
    if (!isActive) {
      return renderFinalLine(line);
    }

    const text = line.content.slice(0, charIndex);
    if (line.type === 'command') {
      return (
        <span>
          <span className="text-[#f5f5f5]">{text}</span>
          <span className="inline-block w-2 h-4 bg-[#22c55e] ml-0.5 animate-pulse align-middle" />
        </span>
      );
    }
    return <span className={getLineClass(line)}>{text}</span>;
  };

  const renderFinalLine = (line: Line) => {
    return <span className={getLineClass(line)}>{line.content}</span>;
  };

  const getLineClass = (line: Line): string => {
    switch (line.type) {
      case 'command':
        return 'text-[#f5f5f5]';
      case 'table':
        return 'text-[#a3a3a3]';
      case 'tree':
        return 'text-[#a3a3a3]';
      default:
        return 'text-[#a3a3a3]';
    }
  };

  const isActiveLine = (index: number) => {
    // The line being typed right now
    if (index < displayLines.length) return false;
    if (index > displayLines.length) return false;
    return index === displayLines.length && currentLine < DEMO_SEQUENCE.length;
  };

  const allLinesToShow = [...displayLines];
  if (currentLine < DEMO_SEQUENCE.length && !isPaused) {
    allLinesToShow.push({
      ...DEMO_SEQUENCE[currentLine],
      content: DEMO_SEQUENCE[currentLine].content.slice(0, currentChar),
    });
  }

  return (
    <div className="w-full max-w-[640px] mx-auto rounded-xl overflow-hidden border border-[#262626] shadow-2xl"
      style={{ boxShadow: '0 0 80px rgba(34, 197, 94, 0.06)' }}
    >
      {/* Window chrome */}
      <div className="flex items-center gap-2 px-4 py-3 bg-[#0c0c0c] border-b border-[#1f1f1f]">
        <div className="w-3 h-3 rounded-full bg-[#ef4444]" />
        <div className="w-3 h-3 rounded-full bg-[#f59e0b]" />
        <div className="w-3 h-3 rounded-full bg-[#22c55e]" />
        <div className="ml-4 font-mono text-xs text-[#525252]">zsh — git-issues demo</div>
      </div>

      {/* Terminal body */}
      <div
        ref={containerRef}
        className="bg-[#0a0a0a] p-5 font-mono text-sm leading-relaxed overflow-y-auto"
        style={{ maxHeight: '420px', minHeight: '380px' }}
      >
        {allLinesToShow.map((line, index) => {
          const active = isActiveLine(index);
          return (
            <div key={`${index}-${line.content}`} className="whitespace-pre">
              {renderContent(line, active, currentChar)}
            </div>
          );
        })}
      </div>
    </div>
  );
}
