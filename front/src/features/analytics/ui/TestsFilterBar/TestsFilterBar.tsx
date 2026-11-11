import type { FC } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { Check, ChevronDown, Filter, Search, X } from "lucide-react";
import styles from "./TestsFilterBar.module.css";

export type TestOption = {
  id: string;
  title: string;
  count: number;
};

type Props = {
  tests: TestOption[];
  totalCount: number;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
};

export const TestsFilterBar: FC<Props> = ({
  tests,
  totalCount,
  selectedId,
  onSelect,
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const wrapRef = useRef<HTMLDivElement>(null);

  const selected = selectedId
    ? tests.find((t) => t.id === selectedId) ?? null
    : null;
  const shownCount = selected ? selected.count : totalCount;

  const filteredTests = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return tests;
    return tests.filter((t) => t.title.toLowerCase().includes(q));
  }, [tests, query]);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  useEffect(() => {
    if (!open) setQuery("");
  }, [open]);

  const pick = (id: string | null) => {
    onSelect(id);
    setOpen(false);
  };

  return (
    <div className={styles.bar}>
      <div className={styles.label}>
        <Filter size={14} strokeWidth={2.2} aria-hidden="true" />
        <span>Фильтр по тесту</span>
      </div>

      <div className={styles.controls}>
        <div className={styles.dropdown} ref={wrapRef}>
          <button
            type="button"
            className={`${styles.trigger} ${selected ? styles.triggerActive : ""}`}
            onClick={() => setOpen((v) => !v)}
            aria-haspopup="listbox"
            aria-expanded={open}
          >
            <span className={styles.triggerLabel} title={selected?.title}>
              {selected ? selected.title : "Все тесты"}
            </span>
            <span className={styles.triggerCount}>{shownCount}</span>
            <ChevronDown
              size={14}
              strokeWidth={2.4}
              className={`${styles.chevron} ${open ? styles.chevronOpen : ""}`}
              aria-hidden="true"
            />
          </button>

          {open && (
            <div className={styles.panel} role="listbox">
              {tests.length > 6 && (
                <div className={styles.searchWrap}>
                  <Search size={13} strokeWidth={2.2} aria-hidden="true" />
                  <input
                    className={styles.searchInput}
                    placeholder="Найти тест..."
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    autoFocus
                  />
                </div>
              )}

              <div className={styles.options}>
                <button
                  type="button"
                  role="option"
                  aria-selected={selectedId === null}
                  className={`${styles.option} ${selectedId === null ? styles.optionActive : ""}`}
                  onClick={() => pick(null)}
                >
                  <span className={styles.optionText}>Все тесты</span>
                  <span className={styles.optionCount}>{totalCount}</span>
                  {selectedId === null && (
                    <Check size={14} strokeWidth={2.6} aria-hidden="true" />
                  )}
                </button>

                {filteredTests.length === 0 && (
                  <p className={styles.noResults}>Ничего не найдено</p>
                )}

                {filteredTests.map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    role="option"
                    aria-selected={selectedId === t.id}
                    className={`${styles.option} ${selectedId === t.id ? styles.optionActive : ""}`}
                    onClick={() => pick(t.id)}
                  >
                    <span className={styles.optionText} title={t.title}>
                      {t.title}
                    </span>
                    <span className={styles.optionCount}>{t.count}</span>
                    {selectedId === t.id && (
                      <Check size={14} strokeWidth={2.6} aria-hidden="true" />
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>

        {selectedId && (
          <button
            type="button"
            className={styles.clearButton}
            onClick={() => onSelect(null)}
            aria-label="Сбросить фильтр"
            title="Сбросить фильтр"
          >
            <X size={14} strokeWidth={2.4} />
          </button>
        )}
      </div>
    </div>
  );
};