import "./globals.css";

export const metadata = {
  title: "DentalDesk AI",
  description: "AI call answering for dental clinics.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
