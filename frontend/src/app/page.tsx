'use client';

import { useState } from 'react';
import { Upload, Package, Printer, CheckCircle } from 'lucide-react';
import Link from 'next/link';

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800">
      {/* Navigation */}
      <nav className="bg-white dark:bg-gray-800 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-blue-600">SwarmForge</h1>
            </div>
            <div className="flex space-x-4">
              <Link href="/dashboard" className="text-gray-700 dark:text-gray-300 hover:text-blue-600">
                Dashboard
              </Link>
              <Link href="/jobs" className="text-gray-700 dark:text-gray-300 hover:text-blue-600">
                Jobs
              </Link>
              <Link href="/printers" className="text-gray-700 dark:text-gray-300 hover:text-blue-600">
                Printers
              </Link>
              <Link href="/login" className="btn-primary">
                Login
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
        <div className="text-center">
          <h1 className="text-5xl font-extrabold text-gray-900 dark:text-white sm:text-6xl">
            Decentralized 3D Printing Network
          </h1>
          <p className="mt-6 text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            Print faster by distributing your 3D models across a network of certified printers. 
            Split large objects into parts and manufacture them simultaneously.
          </p>
          <div className="mt-10 flex justify-center gap-4">
            <Link href="/jobs/new" className="btn-primary text-lg px-8 py-3">
              Submit a Job
            </Link>
            <Link href="/printers/register" className="btn-secondary text-lg px-8 py-3">
              Register Printer
            </Link>
          </div>
        </div>

        {/* Features */}
        <div className="mt-24 grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="card text-center">
            <div className="flex justify-center mb-4">
              <Upload className="w-12 h-12 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold mb-2">Upload & Partition</h3>
            <p className="text-gray-600 dark:text-gray-400">
              Upload your 3D model and our system automatically splits it into optimal parts for distributed printing.
            </p>
          </div>

          <div className="card text-center">
            <div className="flex justify-center mb-4">
              <Printer className="w-12 h-12 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold mb-2">Certified Network</h3>
            <p className="text-gray-600 dark:text-gray-400">
              Print on verified Bambu and Prusa machines with guaranteed quality through our tier-based certification system.
            </p>
          </div>

          <div className="card text-center">
            <div className="flex justify-center mb-4">
              <CheckCircle className="w-12 h-12 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold mb-2">QA Verification</h3>
            <p className="text-gray-600 dark:text-gray-400">
              AI-powered quality assurance with photo verification and dimensional checks ensures perfect results.
            </p>
          </div>
        </div>

        {/* Printer Tiers */}
        <div className="mt-24">
          <h2 className="text-3xl font-bold text-center mb-12">Printer Certification Tiers</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="card border-2 border-purple-300">
              <div className="badge badge-platinum mb-4">PLATINUM</div>
              <h3 className="text-2xl font-bold mb-2">Premium Quality</h3>
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                Industrial-grade printers with ±0.05mm tolerance
              </p>
              <ul className="space-y-2 text-sm">
                <li>✓ Bambu X1 Carbon</li>
                <li>✓ Prusa XL</li>
                <li>✓ Multi-material support</li>
                <li>✓ 95+ calibration score</li>
              </ul>
            </div>

            <div className="card border-2 border-yellow-300">
              <div className="badge badge-gold mb-4">GOLD</div>
              <h3 className="text-2xl font-bold mb-2">Professional Grade</h3>
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                High-quality prosumer printers with ±0.08mm tolerance
              </p>
              <ul className="space-y-2 text-sm">
                <li>✓ Bambu P1S/P1P</li>
                <li>✓ Prusa MK4/MK3S+</li>
                <li>✓ Reliable materials</li>
                <li>✓ 90+ calibration score</li>
              </ul>
            </div>

            <div className="card border-2 border-gray-300">
              <div className="badge badge-silver mb-4">SILVER</div>
              <h3 className="text-2xl font-bold mb-2">Consumer Quality</h3>
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                Entry-level certified printers with ±0.1mm tolerance
              </p>
              <ul className="space-y-2 text-sm">
                <li>✓ Bambu A1</li>
                <li>✓ Prusa Mini+</li>
                <li>✓ Basic materials</li>
                <li>✓ 85+ calibration score</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
