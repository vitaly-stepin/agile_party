import type {
  ServerEvent,
  ClientEvent,
  VotePayload,
  UpdateNicknamePayload,
  SetTaskPayload,
} from '../types';

const WS_BASE_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting';

interface WebSocketClientOptions {
  roomId: string;
  userId: string;
  nickname: string;
  onMessage: (event: ServerEvent) => void;
  onStateChange?: (state: ConnectionState) => void;
  maxReconnectAttempts?: number;
  reconnectInterval?: number;
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private roomId: string;
  private userId: string;
  private nickname: string;
  private onMessage: (event: ServerEvent) => void;
  private onStateChange?: (state: ConnectionState) => void;
  private state: ConnectionState = 'disconnected';
  private reconnectAttempts = 0;
  private maxReconnectAttempts: number;
  private reconnectInterval: number;
  private reconnectTimer: number | null = null;
  private shouldReconnect = true;

  constructor(options: WebSocketClientOptions) {
    this.roomId = options.roomId;
    this.userId = options.userId;
    this.nickname = options.nickname;
    this.onMessage = options.onMessage;
    this.onStateChange = options.onStateChange;
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 10;
    this.reconnectInterval = options.reconnectInterval ?? 1000;
  }

  /**
   * Connect to the WebSocket server
   */
  connect(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.warn('WebSocket is already connected');
      return;
    }

    this.updateState('connecting');
    this.shouldReconnect = true;

    const wsUrl = `${WS_BASE_URL}/ws/rooms/${this.roomId}?userId=${this.userId}&nickname=${encodeURIComponent(this.nickname)}`;

    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.updateState('connected');
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as ServerEvent;
          this.onMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.updateState('disconnected');

        // Attempt to reconnect if not manually closed
        if (this.shouldReconnect && event.code !== 1000) {
          this.attemptReconnect();
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.updateState('disconnected');
      this.attemptReconnect();
    }
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect(): void {
    this.shouldReconnect = false;

    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }

    this.updateState('disconnected');
  }

  /**
   * Submit a vote
   */
  vote(value: string): void {
    this.send({
      type: 'vote',
      payload: { value } as VotePayload,
    });
  }

  /**
   * Reveal votes
   */
  reveal(): void {
    this.send({
      type: 'reveal',
      payload: {},
    });
  }

  /**
   * Clear votes (start new round)
   */
  clear(): void {
    this.send({
      type: 'clear',
      payload: {},
    });
  }

  /**
   * Update user nickname
   */
  updateNickname(nickname: string): void {
    this.send({
      type: 'update_nickname',
      payload: { nickname } as UpdateNicknamePayload,
    });
  }

  /**
   * Set task description
   */
  setTask(description: string): void {
    this.send({
      type: 'set_task',
      payload: { description } as SetTaskPayload,
    });
  }

  /**
   * Get current connection state
   */
  getState(): ConnectionState {
    return this.state;
  }

  /**
   * Send a message to the server
   */
  send(event: ClientEvent): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('Cannot send message: WebSocket is not connected');
      return;
    }

    try {
      this.ws.send(JSON.stringify(event));
    } catch (error) {
      console.error('Failed to send WebSocket message:', error);
    }
  }

  /**
   * Attempt to reconnect with exponential backoff
   */
  private attemptReconnect(): void {
    if (!this.shouldReconnect) {
      return;
    }

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached');
      this.updateState('disconnected');
      return;
    }

    this.reconnectAttempts++;
    this.updateState('reconnecting');

    // Exponential backoff with jitter
    const delay = Math.min(
      this.reconnectInterval * Math.pow(2, this.reconnectAttempts - 1) + Math.random() * 1000,
      30000 // Max 30 seconds
    );

    console.log(`Reconnecting in ${Math.round(delay / 1000)}s (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    this.reconnectTimer = window.setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Update connection state and notify listeners
   */
  private updateState(state: ConnectionState): void {
    this.state = state;
    this.onStateChange?.(state);
  }
}
