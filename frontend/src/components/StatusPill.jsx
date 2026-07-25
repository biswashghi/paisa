export default function StatusPill({ value }) {
  const normalized = String(value).toLowerCase();
  const tone = normalized.includes("active") || normalized.includes("posted") || normalized.includes("published")
    ? "good"
    : normalized.includes("pending") || normalized.includes("draft") || normalized.includes("processing")
      ? "warn"
      : "neutral";
  return <span className={`status-pill ${tone}`}>{value}</span>;
}
