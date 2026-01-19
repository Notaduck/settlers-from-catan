import React from 'react';
// TODO: Implement IncomingTradeModal for responding to player trade offers. Stub UI for now
export function IncomingTradeModal(props: {
  open: boolean;
  onAccept: () => void;
  onDecline: () => void;
  fromPlayer: string;
  offer: Record<string, number>;
  request: Record<string, number>;
}) {
  if (!props.open) return null;
  return (
    <div className="modal" data-cy="incoming-trade-modal">
      <h3>{props.fromPlayer} offers to trade</h3>
      {/* Show offer/request summary */}
      <button data-cy="accept-trade-btn" onClick={props.onAccept}>
        Accept
      </button>
      <button data-cy="decline-trade-btn" onClick={props.onDecline}>
        Decline
      </button>
    </div>
  );
}
