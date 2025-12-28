import { useEffect, useRef, useCallback, useState } from 'react';
import { WebSocketClient } from '../services/websocket';
import type { ConnectionState } from '../services/websocket';
import type { ServerEvent, ClientEvent, Task, TaskListSyncPayload } from '../types';
import { useRoom } from '../context/RoomContext';
import { useTasks } from '../context/TaskContext';
import { api } from '../services/api';

interface UseWebSocketReturn {
  isConnected: boolean;
  connectionState: ConnectionState;
  sendEvent: (event: ClientEvent) => void;
  disconnect: () => void;
}

export const useWebSocket = (roomId: string): UseWebSocketReturn => {
  const { currentUserId, currentUser, setRoomState, updateVotes, setRevealed, updateUserVoteStatus } = useRoom();
  const taskContext = useTasks();
  const tasks = taskContext?.tasks || [];
  const activeTask = taskContext?.activeTask || null;
  const setTasks = taskContext?.setTasks;
  const addTask = taskContext?.addTask;
  const updateTask = taskContext?.updateTask;
  const removeTask = taskContext?.removeTask;
  const reorderTasks = taskContext?.reorderTasks;
  const setActiveTask = taskContext?.setActiveTask;
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const wsClient = useRef<WebSocketClient | null>(null);

  // Use refs to store latest callback values without triggering re-renders
  const setRoomStateRef = useRef(setRoomState);
  const updateVotesRef = useRef(updateVotes);
  const setRevealedRef = useRef(setRevealed);
  const updateUserVoteStatusRef = useRef(updateUserVoteStatus);
  const setTasksRef = useRef(setTasks);
  const addTaskRef = useRef(addTask);
  const updateTaskRef = useRef(updateTask);
  const removeTaskRef = useRef(removeTask);
  const reorderTasksRef = useRef(reorderTasks);
  const setActiveTaskRef = useRef(setActiveTask);
  const tasksRef = useRef(tasks);
  const activeTaskRef = useRef(activeTask);

  // Keep refs in sync with latest values
  useEffect(() => {
    setRoomStateRef.current = setRoomState;
    updateVotesRef.current = updateVotes;
    setRevealedRef.current = setRevealed;
    updateUserVoteStatusRef.current = updateUserVoteStatus;
    setTasksRef.current = setTasks;
    addTaskRef.current = addTask;
    updateTaskRef.current = updateTask;
    removeTaskRef.current = removeTask;
    reorderTasksRef.current = reorderTasks;
    setActiveTaskRef.current = setActiveTask;
    tasksRef.current = tasks;
    activeTaskRef.current = activeTask;
  }, [setRoomState, updateVotes, setRevealed, updateUserVoteStatus, setTasks, addTask, updateTask, removeTask, reorderTasks, setActiveTask, tasks, activeTask]);

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
          // Clear votes for new round
          // Backend will send updated room_state right after this event
          updateVotesRef.current([]);
          setRevealedRef.current(false);
          break;
        }

        case 'error': {
          // Handle error
          const { message } = event.payload;
          console.error('WebSocket error:', message);
          break;
        }

        case 'task_list_sync': {
          const payload = event.payload as TaskListSyncPayload;
          const currentActiveTaskId = activeTaskRef.current?.id;
          setTasksRef.current?.(payload.tasks);

          // If we had an active task, update it with the latest data from the synced list
          // This ensures the active task reference has updated estimation after reveal/clear
          if (currentActiveTaskId && setActiveTaskRef.current) {
            const updatedActiveTask = payload.tasks.find(t => t.id === currentActiveTaskId);
            if (updatedActiveTask) {
              setActiveTaskRef.current(updatedActiveTask);
            }
          }
          break;
        }

        case 'task_created': {
          const task = event.payload as Task;
          addTaskRef.current?.(task);
          break;
        }

        case 'task_updated': {
          const task = event.payload as Task;
          updateTaskRef.current?.(task);
          break;
        }

        case 'task_deleted': {
          const { taskId } = event.payload as { taskId: string };
          removeTaskRef.current?.(taskId);
          break;
        }

        case 'tasks_reordered': {
          const { taskIds } = event.payload as { taskIds: string[] };
          reorderTasksRef.current?.(taskIds);
          break;
        }

        case 'active_task_set': {
          const { taskId } = event.payload as { taskId: string };
          // Find task in current list and set as active
          const task = tasksRef.current.find(t => t.id === taskId);
          if (task && setActiveTaskRef.current) {
            setActiveTaskRef.current(task);
          } else if (taskId && !task) {
            // Task not found in current list, might arrive in next task_list_sync
            // Set a timeout to retry after a brief delay
            setTimeout(() => {
              const retryTask = tasksRef.current.find(t => t.id === taskId);
              if (retryTask && setActiveTaskRef.current) {
                setActiveTaskRef.current(retryTask);
              }
            }, 100);
          }
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
    if (wsClient.current?.send) {
      wsClient.current.send(event);
    }
  }, []);

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
