import { useEffect, useRef, useCallback, useState } from 'react';
import { WebSocketClient } from '../services/websocket';
import type { ConnectionState } from '../services/websocket';
import type { ServerEvent, ClientEvent } from '../types';
import { useRoom } from '../context/RoomContext';
import { api } from '../services/api';

interface UseWebSocketReturn {
  isConnected: boolean;
  connectionState: ConnectionState;
  sendEvent: (event: ClientEvent) => void;
  disconnect: () => void;
}

export const useWebSocket = (roomId: string): UseWebSocketReturn => {
  const { currentUserId, currentUser, setRoomState, updateVotes, setRevealed } = useRoom();
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const wsClient = useRef<WebSocketClient | null>(null);

  const fetchRoomState = useCallback(async () => {
    if (!roomId) return;
    try {
      const state = await api.getRoomState(roomId);
      setRoomState(state);
    } catch (error) {
      console.error('Failed to fetch room state:', error);
    }
  }, [roomId, setRoomState]);

  const handleMessage = useCallback(
    (event: ServerEvent) => {
      console.log('Received WebSocket event:', event.type, event.payload);

      switch (event.type) {
        case 'room_state': {
          // Initial room state sync
          setRoomState(event.payload);
          break;
        }

        case 'user_joined':
        case 'user_left':
        case 'vote_submitted':
        case 'user_updated': {
          // Refresh room state from server
          console.log('User event received:', event.type);
          fetchRoomState();
          break;
        }

        case 'votes_revealed': {
          // Show votes and average
          const { votes, average } = event.payload;
          updateVotes(votes);
          setRevealed(true, average);
          break;
        }

        case 'votes_cleared': {
          // Clear votes for new round
          updateVotes([]);
          setRevealed(false);
          break;
        }

        case 'error': {
          // Handle error
          const { message } = event.payload;
          console.error('WebSocket error:', message);
          break;
        }

        default:
          console.warn('Unknown WebSocket event type:', event.type);
      }
    },
    [setRoomState, updateVotes, setRevealed, fetchRoomState]
  );

  const handleStateChange = useCallback((state: ConnectionState) => {
    console.log('WebSocket state changed:', state);
    setConnectionState(state);
  }, []);

  useEffect(() => {
    if (!roomId || !currentUserId || !currentUser) {
      return;
    }

    // Create WebSocket client
    wsClient.current = new WebSocketClient({
      roomId,
      userId: currentUserId,
      nickname: currentUser.name,
      onMessage: handleMessage,
      onStateChange: handleStateChange,
      maxReconnectAttempts: 10,
      reconnectInterval: 1000,
    });

    // Connect
    wsClient.current.connect();

    // Cleanup on unmount
    return () => {
      if (wsClient.current) {
        wsClient.current.disconnect();
        wsClient.current = null;
      }
    };
  }, [roomId, currentUserId, currentUser, handleMessage, handleStateChange]);

  const sendEvent = useCallback((event: ClientEvent) => {
    if (wsClient.current && connectionState === 'connected') {
      wsClient.current.send(event);
    }
  }, [connectionState]);

  const disconnect = useCallback(() => {
    if (wsClient.current) {
      wsClient.current.disconnect();
    }
  }, []);

  return {
    isConnected: connectionState === 'connected',
    connectionState,
    sendEvent,
    disconnect,
  };
};
