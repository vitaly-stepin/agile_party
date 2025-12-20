import { useState } from 'react';
import { Sparkles, Users, Zap } from 'lucide-react';
import { Card } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import RoomCreation from '../components/room/RoomCreation';
import RoomJoin from '../components/room/RoomJoin';

export default function Home() {
  const [activeTab, setActiveTab] = useState<'create' | 'join'>('create');

  return (
    <div className="min-h-screen relative overflow-hidden flex flex-col items-center justify-center px-4 py-12">
      {/* Animated gradient background */}
      <div className="absolute inset-0 bg-gradient-to-br from-violet-50 via-white to-blue-50" />

      {/* Dot pattern overlay */}
      <div
        className="absolute inset-0 opacity-40"
        style={{
          backgroundImage: `radial-gradient(circle, #e2e8f0 1px, transparent 1px)`,
          backgroundSize: '32px 32px'
        }}
      />

      {/* Gradient orbs */}
      <div className="absolute top-0 -left-4 w-72 h-72 bg-purple-300 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob" />
      <div className="absolute top-0 -right-4 w-72 h-72 bg-blue-300 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob animation-delay-2000" />
      <div className="absolute -bottom-8 left-20 w-72 h-72 bg-violet-300 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob animation-delay-4000" />

      <div className="w-full max-w-2xl relative z-10">
        {/* Hero Section */}
        <div className="text-center mb-12">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-gradient-to-r from-violet-50 to-blue-50 text-violet-700 text-sm font-medium mb-6 border border-violet-200/50 shadow-sm">
            <Sparkles className="w-4 h-4 text-violet-500" />
            Free & No Signup Required
          </div>

          <h1 className="text-6xl font-bold bg-gradient-to-r from-violet-600 via-purple-600 to-blue-600 bg-clip-text text-transparent mb-4 tracking-tight">
            Agile Poker
          </h1>
          <p className="text-xl text-slate-600 max-w-xl mx-auto leading-relaxed">
            Real-time Scrum estimation made simple. Create a room, invite your team, and start estimating tasks together.
          </p>

          {/* Feature Pills */}
          <div className="flex items-center justify-center gap-6 mt-8 text-sm text-slate-600">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-violet-100 to-purple-100 flex items-center justify-center">
                <Users className="w-4 h-4 text-violet-600" />
              </div>
              <span>Unlimited participants</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-100 to-cyan-100 flex items-center justify-center">
                <Zap className="w-4 h-4 text-blue-600" />
              </div>
              <span>Instant updates</span>
            </div>
          </div>
        </div>

        {/* Main Card with Tabs - Gradient border effect */}
        <div className="relative group">
          <div className="absolute -inset-0.5 bg-gradient-to-r from-violet-600 to-blue-600 rounded-xl blur opacity-20 group-hover:opacity-30 transition duration-500" />
          <Card className="relative border-0 shadow-2xl shadow-violet-200/50 bg-white/80 backdrop-blur-sm">
            <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'create' | 'join')}>
              <div className="border-b border-slate-100 px-6 pt-6">
                <TabsList className="w-full grid grid-cols-2 bg-gradient-to-r from-violet-50/50 to-blue-50/50 border border-violet-100/50">
                  <TabsTrigger value="create">
                    Create Room
                  </TabsTrigger>
                  <TabsTrigger value="join">
                    Join Room
                  </TabsTrigger>
                </TabsList>
              </div>

              <div className="p-8">
                <TabsContent value="create">
                  <RoomCreation />
                </TabsContent>

                <TabsContent value="join">
                  <RoomJoin />
                </TabsContent>
              </div>
            </Tabs>
          </Card>
        </div>

        {/* Footer */}
        <p className="text-center text-sm text-slate-500 mt-8">
          Built for agile teams to estimate tasks together
        </p>
      </div>
    </div>
  );
}
