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
  const { currentUserId, currentUser, setRoomState, updateVotes, setRevealed, updateUserVoteStatus } = useRoom();
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const wsClient = useRef<WebSocketClient | null>(null);

  // Use refs to store latest callback values without triggering re-renders
  const setRoomStateRef = useRef(setRoomState);
  const updateVotesRef = useRef(updateVotes);
  const setRevealedRef = useRef(setRevealed);
  const updateUserVoteStatusRef = useRef(updateUserVoteStatus);

  // Keep refs in sync with latest values
  useEffect(() => {
    setRoomStateRef.current = setRoomState;
    updateVotesRef.current = updateVotes;
    setRevealedRef.current = setRevealed;
    updateUserVoteStatusRef.current = updateUserVoteStatus;
  }, [setRoomState, updateVotes, setRevealed, updateUserVoteStatus]);

  const fetchRoomState = useCallback(async () => {
    if (!roomId) return;
    try {
      const state = await api.getRoomState(roomId);
      setRoomStateRef.current(state);
    } catch (error) {
      console.error('Failed to fetch room state:', error);
    }
  }, [roomId]);

  const handleMessage = useCallback(
    (event: ServerEvent) => {
      console.log('Received WebSocket event:', event.type, event.payload);

      switch (event.type) {
        case 'room_state': {
          // Initial room state sync
          setRoomStateRef.current(event.payload);
          break;
        }

        case 'user_joined':
        case 'user_left':
        case 'user_updated': {
          // Refresh room state from server
          console.log('User event received:', event.type);
          fetchRoomState();
          break;
        }

        case 'vote_submitted': {
          // Update voting status immediately from event payload
          const { userId, hasVoted } = event.payload;
          console.log(`Vote submitted by user ${userId}, hasVoted: ${hasVoted}`);

          // Update user's vote status immediately without API call
          updateUserVoteStatusRef.current(userId, hasVoted);
          break;
        }

        case 'votes_revealed': {
          // Show votes and average
          const { votes, average } = event.payload;
          console.log('Votes revealed:', { votes, average });
          updateVotesRef.current(votes);
          setRevealedRef.current(true, average);
          break;
        }

        case 'votes_cleared': {
          // Clear votes for new round and fetch updated user states
          updateVotesRef.current([]);
          setRevealedRef.current(false);
          // Fetch room state to get updated user voting status (all reset to false)
          fetchRoomState();
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
    [fetchRoomState]
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
    // Only reconnect when roomId, userId, or nickname changes - not on every render
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roomId, currentUserId, currentUser?.name]);

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
