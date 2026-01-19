import React from 'react';
// TODO: Implement BankTradeModal for trading with the bank 4:1; stub UI for now
export function BankTradeModal(props: {
  open: boolean;
  onClose: () => void;
  onSubmit: (offering: string, requested: string) => void;
  resources: Record<string, number>;
}) {
  if (!props.open) return null;
  return (
    <div className="modal" data-cy="bank-trade-modal">
      <h3>Bank Trade (4:1)</h3>
      {/* Offer/receive selection UI here */}
      <button data-cy="bank-trade-submit-btn" onClick={() => props.onSubmit('wood', 'brick')}>
        Trade (stub)
      </button>
      <button data-cy="bank-trade-cancel-btn" onClick={props.onClose}>Cancel</button>
    </div>
  );
}
