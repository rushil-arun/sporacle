const AnimatedBackground = () => {
  return (
    <div
      className="fixed inset-0 overflow-hidden pointer-events-none select-none"
      aria-hidden="true"
    >
      <style>{`
        @keyframes flow-a { from { stroke-dashoffset: 2600; } to { stroke-dashoffset: -200; } }
        @keyframes flow-b { from { stroke-dashoffset: 2800; } to { stroke-dashoffset: -200; } }
        @keyframes flow-c { from { stroke-dashoffset: 3200; } to { stroke-dashoffset: -200; } }
        @keyframes flow-d { from { stroke-dashoffset: 2200; } to { stroke-dashoffset: -200; } }

        @keyframes drift-up   { 0%,100% { transform: translateY(0px);  } 50% { transform: translateY(-28px); } }
        @keyframes drift-down { 0%,100% { transform: translateY(0px);  } 50% { transform: translateY(24px);  } }
        @keyframes drift-side { 0%,100% { transform: translateX(0px);  } 50% { transform: translateX(18px);  } }
      `}</style>

      <svg
        viewBox="0 0 1440 900"
        preserveAspectRatio="xMidYMid slice"
        className="absolute inset-0 w-full h-full"
      >
        <g style={{ animation: 'drift-up 20s ease-in-out infinite' }}>
          <path
            d="M -200 210 C 150 90, 480 370, 780 200 C 1080 30, 1300 280, 1640 170"
            fill="none"
            stroke="hsl(255 72% 68%)"
            strokeWidth="1.2"
            strokeOpacity="0.45"
            strokeDasharray="780 1820"
            style={{ animation: 'flow-a 13s linear infinite' }}
          />
          <path
            d="M -200 210 C 150 90, 480 370, 780 200 C 1080 30, 1300 280, 1640 170"
            fill="none"
            stroke="hsl(255 72% 68%)"
            strokeWidth="0.6"
            strokeOpacity="0.20"
            strokeDasharray="1400 1200"
            style={{ animation: 'flow-b 18s linear infinite -6s' }}
          />
        </g>

        <g style={{ animation: 'drift-up 24s ease-in-out infinite -5s' }}>
          <path
            d="M -200 460 C 280 300, 620 620, 940 430 C 1200 270, 1380 500, 1640 360"
            fill="none"
            stroke="hsl(210 88% 64%)"
            strokeWidth="1.4"
            strokeOpacity="0.40"
            strokeDasharray="820 1980"
            style={{ animation: 'flow-b 16s linear infinite -3s' }}
          />
          <path
            d="M -200 460 C 280 300, 620 620, 940 430 C 1200 270, 1380 500, 1640 360"
            fill="none"
            stroke="hsl(210 88% 64%)"
            strokeWidth="0.5"
            strokeOpacity="0.18"
            strokeDasharray="1600 1400"
            style={{ animation: 'flow-a 22s linear infinite -10s' }}
          />
        </g>

        <g style={{ animation: 'drift-down 19s ease-in-out infinite -8s' }}>
          <path
            d="M -200 690 C 300 580, 640 790, 960 650 C 1260 510, 1400 730, 1640 620"
            fill="none"
            stroke="hsl(322 72% 62%)"
            strokeWidth="1.3"
            strokeOpacity="0.42"
            strokeDasharray="760 2040"
            style={{ animation: 'flow-c 15s linear infinite -7s' }}
          />
          <path
            d="M -200 690 C 300 580, 640 790, 960 650 C 1260 510, 1400 730, 1640 620"
            fill="none"
            stroke="hsl(322 72% 62%)"
            strokeWidth="0.5"
            strokeOpacity="0.18"
            strokeDasharray="1300 1700"
            style={{ animation: 'flow-d 20s linear infinite -2s' }}
          />
        </g>

        <g style={{ animation: 'drift-side 22s ease-in-out infinite -4s' }}>
          <path
            d="M 340 -80 C 220 180, 460 420, 320 640 C 200 830, 440 870, 380 1000"
            fill="none"
            stroke="hsl(350 74% 60%)"
            strokeWidth="1.2"
            strokeOpacity="0.38"
            strokeDasharray="650 1850"
            style={{ animation: 'flow-d 17s linear infinite -1s' }}
          />
        </g>

        <g style={{ animation: 'drift-side 18s ease-in-out infinite -11s' }}>
          <path
            d="M 1080 -80 C 1200 200, 1020 420, 1100 660 C 1180 860, 1000 880, 1040 1000"
            fill="none"
            stroke="hsl(265 76% 66%)"
            strokeWidth="1.2"
            strokeOpacity="0.36"
            strokeDasharray="680 1920"
            style={{ animation: 'flow-b 19s linear infinite -9s' }}
          />
        </g>

        <g style={{ animation: 'drift-up 26s ease-in-out infinite -14s' }}>
          <path
            d="M -200 330 C 100 460, 400 180, 720 380 C 1020 560, 1280 280, 1640 440"
            fill="none"
            stroke="hsl(240 78% 68%)"
            strokeWidth="1.0"
            strokeOpacity="0.32"
            strokeDasharray="700 2100"
            style={{ animation: 'flow-a 21s linear infinite -5s' }}
          />
        </g>

        <g style={{ animation: 'drift-down 17s ease-in-out infinite -3s' }}>
          <path
            d="M -200 800 C 400 720, 750 860, 1080 790 C 1300 740, 1450 820, 1640 780"
            fill="none"
            stroke="hsl(320 68% 60%)"
            strokeWidth="1.0"
            strokeOpacity="0.30"
            strokeDasharray="600 2200"
            style={{ animation: 'flow-c 12s linear infinite -8s' }}
          />
        </g>

        <g style={{ animation: 'drift-up 21s ease-in-out infinite -7s' }}>
          <path
            d="M 900 -80 C 1050 100, 800 280, 1000 420 C 1150 520, 1350 380, 1640 500"
            fill="none"
            stroke="hsl(210 88% 64%)"
            strokeWidth="1.1"
            strokeOpacity="0.30"
            strokeDasharray="580 2020"
            style={{ animation: 'flow-d 14s linear infinite -12s' }}
          />
        </g>
      </svg>
    </div>
  );
};

export default AnimatedBackground;
