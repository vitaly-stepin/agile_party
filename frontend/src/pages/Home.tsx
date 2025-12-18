import { useState } from 'react';
import { Card } from '../components/common';
import RoomCreation from '../components/room/RoomCreation';
import RoomJoin from '../components/room/RoomJoin';

export default function Home() {
  const [activeTab, setActiveTab] = useState<'create' | 'join'>('create');

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Agile Poker
          </h1>
          <p className="text-gray-600">
            Real-time Scrum estimation made simple
          </p>
        </div>

        <Card variant="elevated" padding="lg">
          {/* Tabs */}
          <div className="flex border-b border-gray-200 mb-6">
            <button
              onClick={() => setActiveTab('create')}
              className={`flex-1 py-3 px-4 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'create'
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Create Room
            </button>
            <button
              onClick={() => setActiveTab('join')}
              className={`flex-1 py-3 px-4 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'join'
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Join Room
            </button>
          </div>

          {/* Tab Content */}
          {activeTab === 'create' ? <RoomCreation /> : <RoomJoin />}
        </Card>

        <p className="text-center text-sm text-gray-500 mt-6">
          Built for agile teams to estimate tasks together
        </p>
      </div>
    </div>
  );
}
