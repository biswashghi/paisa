export default function PaisaLogo({ variant = "compact" }) {
  return (
    <div className={`paisa-logo ${variant}`}>
      <img src="/paisa-wordmark.png" alt="Paisa" />
    </div>
  );
}
