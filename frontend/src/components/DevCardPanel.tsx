import React, { useState } from "react";
import { useGame } from "@/context/GameContext";
import { DevCardType, Resource } from "@/types";

const CARD_TYPES = [
  {
    type: DevCardType.KNIGHT,
    name: "Knight",
    desc: "Move the robber and steal a resource. Counts for Largest Army.",
  },
  {
    type: DevCardType.ROAD_BUILDING,
    name: "Road Building",
    desc: "Place 2 free roads.",
  },
  {
    type: DevCardType.YEAR_OF_PLENTY,
    name: "Year of Plenty",
    desc: "Take any 2 resources from the bank.",
  },
  {
    type: DevCardType.MONOPOLY,
    name: "Monopoly",
    desc: "Choose a resource. All other players give you all of that type.",
  },
  {
    type: DevCardType.VICTORY_POINT,
    name: "Victory Point",
    desc: "Counts toward victory. Hidden until you win or reveal.",
  },
];

const RESOURCE_TYPES = [
  Resource.WOOD,
  Resource.BRICK,
  Resource.SHEEP,
  Resource.WHEAT,
  Resource.ORE,
];

const RESOURCE_LABELS: Record<Resource, string> = {
  [Resource.WOOD]: "Wood",
  [Resource.BRICK]: "Brick",
  [Resource.SHEEP]: "Sheep",
  [Resource.WHEAT]: "Wheat",
  [Resource.ORE]: "Ore",
  [Resource.UNSPECIFIED]: "?",
};

export function DevCardPanel() {
  const {
    gameState,
    currentPlayerId,
    buyDevCard,
    playDevCard,
    error,
  } = useGame();
  const [modal, setModal] = useState<null | { type: "monopoly" | "year", onPick: (resource: Resource | Resource[]) => void }>(null);
  const [pending, setPending] = useState(false);
  const [notification, setNotification] = useState<string | null>(null);

  const me = gameState?.players.find((p) => p.id === currentPlayerId) ?? null;
  const turn = gameState?.turnCounter ?? 0;

  const makeEnumRecord = () => ({
    [DevCardType.UNSPECIFIED]: 0,
    [DevCardType.KNIGHT]: 0,
    [DevCardType.ROAD_BUILDING]: 0,
    [DevCardType.YEAR_OF_PLENTY]: 0,
    [DevCardType.MONOPOLY]: 0,
    [DevCardType.VICTORY_POINT]: 0,
  });

  const purchasedNow = React.useMemo(() => {
    const rec = makeEnumRecord();
    if (!me?.devCardsPurchasedTurn) {
      return rec;
    }
    Object.entries(me.devCardsPurchasedTurn).forEach(([key, val]) => {
      if (val === turn) {
        const type = Number(key) as DevCardType;
        rec[type] = (rec[type] || 0) + 1;
      }
    });
    return rec;
  }, [me, turn]);

  const devCards = me?.devCards ?? {};
  const victoryPoints = me?.victoryPointCards || 0;
  const playableCards: Record<DevCardType, number> = makeEnumRecord();
  Object.entries(devCards).forEach(([key, val]) => {
    const type = Number(key) as DevCardType;
    const total = val as number;
    const bought = purchasedNow[type] || 0;
    if (type !== DevCardType.VICTORY_POINT) {
      playableCards[type] = total - bought;
    }
  });

  const canBuy = Boolean(
    me?.resources &&
      me.resources.ore > 0 &&
      me.resources.wheat > 0 &&
      me.resources.sheep > 0 &&
      (gameState?.devCardDeck?.length ?? 0) > 0
  );

  // Handler: Buy
  function handleBuy() {
    setPending(true);
    buyDevCard();
    setTimeout(() => {
      setPending(false);
      setNotification(null);
    }, 850);
  }

  // Handler: Play card
  function handlePlay(cardType: DevCardType) {
    if (cardType === DevCardType.MONOPOLY) {
      setModal({
        type: "monopoly",
        onPick: (resource) => {
          playDevCard(cardType, resource as Resource);
          setModal(null);
        },
      });
      return;
    }
    if (cardType === DevCardType.YEAR_OF_PLENTY) {
      setModal({
        type: "year",
        onPick: (resources) => {
          if (Array.isArray(resources) && resources.length === 2) {
            playDevCard(cardType, undefined, resources as Resource[]);
          }
          setModal(null);
        },
      });
      return;
    }
    playDevCard(cardType);
  }

  // Show notification on just bought
  React.useEffect(() => {
    if (!me) {
      setNotification(null);
      return;
    }
    const t = Object.keys(purchasedNow)
      .map((k) => CARD_TYPES.find((ct) => ct.type === Number(k))?.name)
      .filter(Boolean)
      .join(", ");
    setNotification(t ? `You drew: ${t}` : null);
  }, [me, turn, purchasedNow]);

  if (!gameState || !currentPlayerId || !me) return null;

  // Data-cy keys per spec
  function dc(type: DevCardType) {
    return `dev-card-${CARD_TYPES.find((ct) => ct.type === type)?.name?.toLowerCase().replace(/\s+/g, '-')}`;
  }
  // Calculate total number of dev cards (including hidden VP)
  const totalDevCards = Object.values(devCards).reduce((sum, n) => sum + (n || 0), 0);
  return (
    <section className="devcard-panel" data-cy="dev-cards-panel">
      <h2>{`Development Cards (${totalDevCards})`}</h2>
      <div className="dev-cards-list devcard-hand">
        {CARD_TYPES.map((card) => {
          // Victory Point: hide from other players (this is your hand) except at game end
          const count = devCards[card.type] || 0;
          if (card.type === DevCardType.VICTORY_POINT && !count) return null;
          return (
            <div
              key={card.type}
              data-cy={dc(card.type)}
              className={`dev-card-item devcard devcard-${card.name.toLowerCase().replace(/\s+/g, '-')}`}
            >
              <span className="dev-card-name devcard-label">{card.name}</span>
              {count > 1 && <span className="dev-card-count">×{count}</span>}
              {/* Only show play for playable cards, and not for VP cards */}
              {playableCards[card.type] > 0 && (
                <button
                  className="play-btn"
                  data-cy={`play-dev-card-btn-${card.name.toLowerCase().replace(/\s+/g, '-')}`}
                  disabled={pending}
                  onClick={() => handlePlay(card.type)}
                >
                  Play
                </button>
              )}
              {/* VP cards: show count if game won, else keep hidden */}
              {card.type === DevCardType.VICTORY_POINT && (
                <span className="devcard-vp">{victoryPoints > 0 ? `+${victoryPoints} VP` : ""}</span>
              )}
            </div>
          );
        })}
      </div>
      <button
        data-cy="buy-dev-card-btn"
        disabled={!canBuy || pending}
        onClick={handleBuy}
        className="buy-btn"
      >
        Buy Development Card
      </button>
      {notification && (
        <div className="devcard-notification">{notification}</div>
      )}
      {error && <div className="devcard-error">{error}</div>}
      {/* Modal: Monopoly pick */}
      {modal?.type === "monopoly" && (
        <div className="modal monopoly-modal" data-cy="monopoly-modal">
          <h3>Monopoly: Pick a resource</h3>
          <div className="monopoly-pick">
            {RESOURCE_TYPES.map((res) => (
              <button
                key={res}
                data-cy={`monopoly-select-${RESOURCE_LABELS[res].toLowerCase()}`}
                onClick={() => modal.onPick(res)}
              >
                {RESOURCE_LABELS[res]}
              </button>
            ))}
          </div>
          <button data-cy="monopoly-submit" disabled>
            Submit
          </button>
        </div>
      )}
      {modal?.type === "year" && (
        <YearOfPlentyModal
          onPick={(selected) => {
            if (selected.length === 2) modal.onPick(selected);
          }}
          onClose={() => setModal(null)}
        />
      )}
    </section>
  );
}

// Modal for Year of Plenty (two picks)
function YearOfPlentyModal({ onPick, onClose }: { onPick: (selected: Resource[]) => void; onClose: () => void }) {
  const [selection, setSelection] = useState<Resource[]>([]);
  function handleSelect(res: Resource) {
    if (selection.length < 2) setSelection((sel) => [...sel, res]);
  }
  function handleReset() {
    setSelection([]);
  }
  return (
    <div className="modal yearmodal" data-cy="year-of-plenty-modal">
      <h3>Year of Plenty: Pick any 2 resources</h3>
      <div className="year-pick">
        {RESOURCE_TYPES.map((res) => (
          <button
            key={res}
            data-cy={`year-of-plenty-select-${RESOURCE_LABELS[res].toLowerCase()}`}
            onClick={() => handleSelect(res)}
            disabled={selection.length >= 2}
          >
            {RESOURCE_LABELS[res]}
          </button>
        ))}
      </div>
      <div className="year-selected">
        Selected: {selection.map((r) => RESOURCE_LABELS[r]).join(", ")}
      </div>
      <button
        data-cy="year-of-plenty-submit"
        onClick={() => selection.length === 2 && onPick(selection)}
        disabled={selection.length !== 2}
      >
        Confirm
      </button>
      <button onClick={handleReset}>Reset</button>
      <button onClick={onClose}>Cancel</button>
    </div>
  );
}

export default DevCardPanel;
