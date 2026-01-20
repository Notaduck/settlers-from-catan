// TODO: Implement ProposeTradeModal for proposing trade to player or all; stub UI for now
export function ProposeTradeModal(props: {
  open: boolean;
  onClose: () => void;
  onSubmit: (offer: Record<string, number>, request: Record<string, number>, toPlayerId?: string|null) => void;
  players: {id: string; name: string;}[];
  myResources: Record<string, number>;
}) {
  if (!props.open) return null;
  return (
    <div className="modal" data-cy="propose-trade-modal">
      <h3>Propose Trade</h3>
      {/* Offer/request selection UI here */}
      <button data-cy="propose-trade-submit-btn" onClick={() => props.onSubmit({wood:1},{brick:1},null)}>
        Propose (stub)
      </button>
      <button data-cy="propose-trade-cancel-btn" onClick={props.onClose}>Cancel</button>
    </div>
  );
}
