import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "SF Trails - Trail Status Checker",
  description: "Check trail status in San Francisco area parks",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased bg-gray-50 min-h-screen`}
      >
        <header className="sticky top-0 z-10 border-b border-gray-200 bg-white">
          <nav className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
            <Link href="/" className="flex items-center gap-2">
              <span className="text-xl font-bold text-green-600">SF Trails</span>
            </Link>
            <div className="flex items-center gap-6">
              <Link
                href="/trails"
                className="text-sm font-medium text-gray-600 hover:text-gray-900"
              >
                All Trails
              </Link>
              <Link
                href="/parks"
                className="text-sm font-medium text-gray-600 hover:text-gray-900"
              >
                Parks
              </Link>
            </div>
          </nav>
        </header>
        <main>{children}</main>
        <footer className="border-t border-gray-200 bg-white mt-auto">
          <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
            <p className="text-center text-sm text-gray-500">
              SF Trails - Trail status information for San Francisco area parks
            </p>
          </div>
        </footer>
      </body>
    </html>
  );
}
